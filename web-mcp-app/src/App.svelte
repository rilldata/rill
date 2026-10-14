<script lang="ts">
  import "./styles/app.css";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { onDestroy, onMount, untrack } from "svelte";
  import ChartCard from "./components/ChartCard.svelte";
  import EmptyState from "./components/EmptyState.svelte";
  import ErrorState from "./components/ErrorState.svelte";
  import FullscreenView from "./components/FullscreenView.svelte";
  import Skeleton from "./components/Skeleton.svelte";
  import { adaptResult, type AdaptedChart } from "./lib/adapt";
  import {
    isFlintResult,
    textContent,
    withChartData,
    type ChartDataPayload,
    type ChartResult,
    type WireChartSpec,
    type WireFlintChartSpec,
    type WireFlintInput,
  } from "./lib/contract";
  import {
    controlOptions,
    describeView,
    patchSpec,
    topNTarget,
    type ControlState,
  } from "./lib/controls";
  import {
    describeFlintView,
    flintControlOptions,
    flintTopNTarget,
    patchFlintInput,
  } from "./lib/flint-controls";
  import { buildFlintPanel } from "./lib/flint-options";
  import { HostSession } from "./lib/host.svelte";
  import { setLocaleFromHost } from "./lib/i18n";
  import { inlineChartHeight } from "./lib/sizing";
  import {
    applyHostTheme,
    SYSTEM_FONT_STACK,
    type AppliedTheme,
  } from "./lib/theme";
  import { interpretToolResult } from "./lib/view";

  type ViewState =
    | { kind: "connecting" }
    | { kind: "loading" }
    | { kind: "error"; message: string; openUrl?: string }
    | { kind: "empty"; openUrl?: string }
    | { kind: "ready"; chart: AdaptedChart };

  const session = new HostSession();

  let applied = $state.raw<AppliedTheme>({
    isDark: false,
    fontFamily: SYSTEM_FONT_STACK,
  });
  let locale = $state("en");

  // Interactive controls: the user's choices and the data they produced.
  // Raw: these objects feed the chart spec and must not be proxied (see HostSession).
  let controls = $state.raw<ControlState>({});
  let dataOverride = $state.raw<
    | { patched: WireChartSpec | WireFlintInput; payload: ChartDataPayload }
    | undefined
  >(undefined);
  // Flint charts: the presentation controls' edit of the chart spec, compiled client-side.
  let presentationSpec = $state.raw<WireFlintChartSpec | undefined>(undefined);
  let refreshing = $state(false);
  let refreshError = $state<string | undefined>(undefined);
  let requestSeq = 0;

  $effect(() => {
    const next = applyHostTheme(session.hostContext);
    // Every host-context change lands here, including resizes; a new `applied` object
    // re-adapts the result and re-embeds the chart, so it only changes when its values do.
    untrack(() => {
      if (
        next.isDark !== applied.isDark ||
        next.fontFamily !== applied.fontFamily
      ) {
        applied = next;
      }
    });
  });
  $effect(() => {
    locale = setLocaleFromHost(session.hostContext?.locale);
  });
  // A new tool result starts over: the edits belonged to the previous one.
  $effect(() => {
    void session.toolResult;
    untrack(() => {
      controls = {};
      dataOverride = undefined;
      presentationSpec = undefined;
      refreshError = undefined;
    });
  });

  onMount(() => {
    void session.connect();
  });
  onDestroy(() => session.destroy());

  const interpreted = $derived(
    session.toolResult ? interpretToolResult(session.toolResult) : undefined,
  );
  const baseResult = $derived(
    interpreted?.kind === "chart" ? interpreted.result : undefined,
  );
  const effectiveResult = $derived.by((): ChartResult | undefined => {
    if (!baseResult) return undefined;
    if (!dataOverride) return baseResult;
    return withChartData(
      baseResult,
      dataOverride.patched,
      dataOverride.payload,
    );
  });
  const timeDimension = $derived(baseResult?.metrics_view?.time_dimension);
  const timeRange = $derived(
    baseResult
      ? (baseResult.resolved_time_range ??
          (isFlintResult(baseResult)
            ? baseResult.input.data.time_range
            : baseResult.spec.time_range))
      : undefined,
  );
  /** The Flint chart spec in effect: the presentation edit, else the (refreshed) result's. */
  const currentFlintSpec = $derived(
    presentationSpec ??
      (effectiveResult && isFlintResult(effectiveResult)
        ? effectiveResult.input.chart_spec
        : undefined),
  );

  const view = $derived.by((): ViewState => {
    switch (session.status) {
      case "connecting":
        return { kind: "connecting" };
      case "awaiting-input":
      case "awaiting-result":
        return { kind: "loading" };
      case "cancelled":
        return { kind: "error", message: m.mcp_chart_error_cancelled() };
      case "failed":
        return {
          kind: "error",
          message: session.error ?? m.mcp_chart_error_tool_failed(),
        };
      case "received":
        break;
    }
    if (!interpreted || !effectiveResult) {
      return interpreted?.kind === "error"
        ? {
            kind: "error",
            message: interpreted.message,
            openUrl: interpreted.openUrl,
          }
        : { kind: "loading" };
    }
    try {
      const chart = adaptResult(effectiveResult, {
        isDark: applied.isDark,
        locale,
        chartSpec: presentationSpec,
      });
      if (chart.rowCount === 0)
        return { kind: "empty", openUrl: chart.openUrl };
      return { kind: "ready", chart };
    } catch (error) {
      console.error("rill chart widget: could not adapt the result", error);
      return {
        kind: "error",
        message: m.mcp_chart_error_render({
          message: error instanceof Error ? error.message : String(error),
        }),
        openUrl: effectiveResult.open_url,
      };
    }
  });

  const chartHeight = $derived(
    inlineChartHeight(session.hostContext?.containerDimensions),
  );

  const availableControls = $derived.by(() => {
    if (!baseResult || !session.canCallServerTools) return undefined;
    const smallestGrain = baseResult.metrics_view?.smallest_time_grain;
    if (isFlintResult(baseResult)) {
      if (!currentFlintSpec) return undefined;
      return flintControlOptions(
        currentFlintSpec,
        baseResult.fields,
        timeDimension,
        timeRange,
        smallestGrain,
      );
    }
    return controlOptions(
      baseResult.chart_type,
      baseResult.spec,
      timeRange,
      smallestGrain,
    );
  });

  // What the controls show as selected: the user's choice, else what the result carries.
  // The result's spec holds the top-N limit the backend applied, defaults and caps included.
  const currentControls = $derived.by((): ControlState => {
    if (!baseResult) return {};
    const topN = isFlintResult(baseResult)
      ? currentFlintSpec &&
        flintTopNTarget(currentFlintSpec, baseResult.fields, timeDimension)
          ?.limit
      : topNTarget(baseResult.chart_type, baseResult.spec)?.limit;
    return {
      grain: controls.grain ?? effectiveResult?.time_grain,
      topN: controls.topN ?? topN,
      comparison:
        controls.comparison ??
        (view.kind === "ready" ? view.chart.hasComparison : false),
    };
  });

  // The presentation controls are built only when they are shown: Flint ranks and test-compiles
  // the candidate chart types, which is cheap but not free.
  const panel = $derived.by(() => {
    if (view.kind !== "ready" || view.chart.engine !== "flint")
      return undefined;
    if (!session.isFullscreen) return undefined;
    // Chart-type choices come from the result's own spec so they stay put across switches.
    const base =
      effectiveResult && isFlintResult(effectiveResult)
        ? effectiveResult.input.chart_spec
        : view.chart.chartSpec;
    return buildFlintPanel(
      view.chart.chartSpec,
      view.chart.flint,
      view.chart.rows,
      base,
    );
  });

  function currentDescription(): string | undefined {
    if (view.kind !== "ready") return undefined;
    const chart = view.chart;
    return chart.engine === "flint"
      ? describeFlintView(
          chart.title,
          chart.chartSpec,
          currentControls,
          chart.timeRangeLabel,
        )
      : describeView(chart.title, currentControls, chart.timeRangeLabel);
  }

  async function applyControls(next: ControlState) {
    if (!baseResult || !effectiveResult) return;
    controls = next;
    // Rill Cloud's unified MCP server routes every call by a `project` argument, which it
    // strips before forwarding; the argument is echoed from the original `create_chart` input.
    const project = session.toolInput?.project;
    let args: Record<string, unknown>;
    let patched: WireChartSpec | WireFlintInput;
    if (isFlintResult(baseResult) && isFlintResult(effectiveResult)) {
      const input = patchFlintInput(
        effectiveResult.input,
        currentFlintSpec ?? effectiveResult.input.chart_spec,
        next,
        timeRange,
        baseResult.fields,
        timeDimension,
      );
      patched = input;
      args = { chart_spec: input.chart_spec, data: input.data };
    } else {
      const spec = patchSpec(
        baseResult.spec as WireChartSpec,
        baseResult.chart_type as never,
        next,
        timeRange,
      );
      patched = spec;
      args = { chart_type: baseResult.chart_type, spec };
    }
    const seq = ++requestSeq;
    refreshing = true;
    refreshError = undefined;
    try {
      const response = await session.callServerTool("chart_data", {
        ...(typeof project === "string" && { project }),
        ...args,
      });
      if (seq !== requestSeq) return;
      if (response.isError) {
        throw new Error(textContent(response) || "chart_data failed");
      }
      const payload = (response.structuredContent ?? {}) as ChartDataPayload;
      dataOverride = { patched, payload };
      // The refreshed input carries the presentation edits the request was built from.
      presentationSpec = undefined;
      // `view` has already been derived from the override at this point.
      const description = currentDescription();
      if (description) session.updateModelContext(description);
    } catch (error) {
      if (seq !== requestSeq) return;
      console.warn("rill chart widget: chart_data call failed", error);
      refreshError = m.mcp_chart_refresh_failed();
    } finally {
      if (seq === requestSeq) refreshing = false;
    }
  }

  function applyPresentation(next: WireFlintChartSpec) {
    presentationSpec = next;
    const description = currentDescription();
    if (description) session.updateModelContext(description);
  }

  function openLink(url: string | undefined) {
    if (url) void session.openLink(url);
  }
</script>

{#if view.kind === "ready" && session.isFullscreen}
  <FullscreenView
    chart={view.chart}
    isDark={applied.isDark}
    fontFamily={applied.fontFamily}
    safeAreaInsets={session.hostContext?.safeAreaInsets}
    controls={availableControls}
    controlState={currentControls}
    {panel}
    {refreshing}
    {refreshError}
    openUrl={session.canOpenLinks ? view.chart.openUrl : undefined}
    onOpenLink={() =>
      openLink(view.kind === "ready" ? view.chart.openUrl : undefined)}
    onExit={() => void session.requestDisplayMode("inline")}
    onControlsChange={(next) => void applyControls(next)}
    onPresentationChange={applyPresentation}
  />
{:else}
  <div class="widget">
    {#if view.kind === "connecting"}
      <Skeleton height={chartHeight} label={m.mcp_chart_connecting()} />
    {:else if view.kind === "loading"}
      <Skeleton height={chartHeight} label={m.mcp_chart_loading()} />
    {:else if view.kind === "error"}
      <ErrorState
        message={view.message}
        openUrl={session.canOpenLinks ? view.openUrl : undefined}
        onOpenLink={() =>
          openLink(view.kind === "error" ? view.openUrl : undefined)}
      />
    {:else if view.kind === "empty"}
      <EmptyState
        openUrl={session.canOpenLinks ? view.openUrl : undefined}
        onOpenLink={() =>
          openLink(view.kind === "empty" ? view.openUrl : undefined)}
      />
    {:else}
      <ChartCard
        chart={view.chart}
        isDark={applied.isDark}
        fontFamily={applied.fontFamily}
        {chartHeight}
        canFullscreen={session.canFullscreen}
        openUrl={session.canOpenLinks ? view.chart.openUrl : undefined}
        {refreshing}
        {refreshError}
        onOpenLink={() =>
          openLink(view.kind === "ready" ? view.chart.openUrl : undefined)}
        onFullscreen={() => void session.requestDisplayMode("fullscreen")}
      />
    {/if}
  </div>
{/if}

<style lang="postcss">
  .widget {
    @apply w-full;
    min-width: 320px;
  }
</style>
