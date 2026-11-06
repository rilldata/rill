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
    isRecord,
    parseChartDataResponse,
    textContent,
    withChartData,
    type ChartDataPayload,
    type ChartDataResponse,
    type ChartResult,
    type WireFlintChartSpec,
    type WireFlintInput,
  } from "./lib/contract";
  import {
    controlOptions,
    describeView,
    patchInput,
    topNTarget,
    type ControlState,
  } from "./lib/controls";
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

  // Interactive controls: the choice of the refresh in flight, and the data the last refresh produced.
  // Once a refresh settles the result in effect is what the controls show, so a failed one rolls them back.
  // Raw: these objects feed the chart spec and must not be proxied (see HostSession).
  let controls = $state.raw<ControlState>({});
  let dataOverride = $state.raw<
    { queried: WireFlintInput; payload: ChartDataPayload } | undefined
  >(undefined);
  // The presentation controls' edit of the chart spec, compiled client-side.
  let presentationSpec = $state.raw<WireFlintChartSpec | undefined>(undefined);
  let refreshing = $state(false);
  let refreshError = $state<string | undefined>(undefined);
  // Every control change re-queries the data through `chart_data`. Changes in quick succession (stepping through
  // the top-N values, say) are coalesced: the request goes out once the controls have been still for a moment,
  // and each change invalidates the in-flight one, whose response is then dropped.
  const refreshDebounceMs = 250;
  let refreshTimer: ReturnType<typeof setTimeout> | undefined;
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
  // A new tool result starts over: the edits belonged to the previous one, and so does any pending refresh.
  $effect(() => {
    void session.toolResult;
    untrack(() => {
      clearTimeout(refreshTimer);
      requestSeq++;
      controls = {};
      dataOverride = undefined;
      presentationSpec = undefined;
      refreshing = false;
      refreshError = undefined;
    });
  });

  onMount(() => {
    void session.connect();
  });
  onDestroy(() => {
    clearTimeout(refreshTimer);
    session.destroy();
  });

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
      dataOverride.queried,
      dataOverride.payload,
    );
  });
  const timeDimension = $derived(baseResult?.metrics_view?.time_dimension);
  const timeRange = $derived(
    baseResult
      ? (baseResult.resolved_time_range ?? baseResult.input.data.time_range)
      : undefined,
  );
  /** The chart spec in effect: the presentation edit, else the (refreshed) result's. */
  const currentSpec = $derived(
    presentationSpec ?? effectiveResult?.input.chart_spec,
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
    if (!baseResult || !currentSpec || !session.canCallServerTools) {
      return undefined;
    }
    return controlOptions(
      currentSpec,
      effectiveResult?.fields,
      timeDimension,
      timeRange,
      baseResult.metrics_view?.smallest_time_grain,
    );
  });

  // What the controls show as selected: the choice in flight, else what the result in effect carries.
  // Its spec holds the top-N limit the backend applied, defaults and caps included.
  const currentControls = $derived.by((): ControlState => {
    if (!effectiveResult || !currentSpec) return {};
    const topN = topNTarget(
      currentSpec,
      effectiveResult.fields,
      timeDimension,
    )?.limit;
    return {
      grain: controls.grain ?? effectiveResult.time_grain,
      topN: controls.topN ?? topN,
      comparison:
        controls.comparison ??
        (view.kind === "ready" ? view.chart.hasComparison : false),
    };
  });

  // The presentation controls are built only when they are shown: reading Flint's options compiles
  // the chart, which is cheap but not free.
  const panel = $derived.by(() => {
    if (view.kind !== "ready" || !session.isFullscreen || !effectiveResult) {
      return undefined;
    }
    // Chart-type choices come from the result's own spec so they stay put across switches.
    return buildFlintPanel(
      view.chart.chartSpec,
      view.chart.flint,
      effectiveResult.fields ?? {},
      effectiveResult.input.chart_spec,
    );
  });

  /** Tells the model what the chart shows now. */
  function updateModelContext() {
    if (view.kind !== "ready") return;
    const chart = view.chart;
    session.updateModelContext(
      describeView(
        chart.title,
        chart.chartSpec,
        currentControls,
        chart.timeRangeLabel,
      ),
    );
  }

  function applyControls(next: ControlState) {
    if (!baseResult) return;
    controls = next;
    refreshing = true;
    refreshError = undefined;
    const seq = ++requestSeq;
    clearTimeout(refreshTimer);
    refreshTimer = setTimeout(
      () => void fetchChartData(next, seq),
      refreshDebounceMs,
    );
  }

  async function fetchChartData(next: ControlState, seq: number) {
    if (!effectiveResult) {
      refreshing = false;
      return;
    }
    // Rill Cloud's unified MCP server routes every call by a `project` argument, which it
    // strips before forwarding; the argument is echoed from the original `create_chart` input.
    const project = session.toolInput?.project;
    const input = patchInput(
      effectiveResult.input,
      currentSpec ?? effectiveResult.input.chart_spec,
      next,
      timeRange,
      effectiveResult.fields,
      timeDimension,
    );
    // The caller's semantic type overrides hold for every query of the chart: the input echoes them,
    // else the original call carries them. Without any the key is left out, since the input schema takes no null.
    const semanticTypes =
      input.semantic_types ?? session.toolInput?.semantic_types;
    let failure: string | undefined;
    try {
      const response = await session.callServerTool("chart_data", {
        ...(typeof project === "string" && { project }),
        chart_spec: input.chart_spec,
        data: input.data,
        ...(isRecord(semanticTypes) &&
          Object.keys(semanticTypes).length > 0 && {
            semantic_types: semanticTypes,
          }),
      });
      if (seq !== requestSeq) return;
      if (response.isError) {
        throw new Error(textContent(response) || "chart_data failed");
      }
      const parsed = parseChartDataResponse(response.structuredContent);
      if (parsed.kind === "invalid") {
        throw new Error(`invalid chart_data result: ${parsed.reason}`);
      }
      failure = applyChartData(parsed.response, input);
    } catch (error) {
      if (seq !== requestSeq) return;
      console.warn("rill chart widget: chart_data call failed", error);
      failure = m.mcp_chart_refresh_failed();
    }
    // Settled: the controls show the state in effect again, which a failure leaves as it was.
    controls = {};
    refreshing = false;
    refreshError = failure;
    updateModelContext();
  }

  /**
   * Applies a refresh, unless the chart cannot show it: `view` derives from the override on read,
   * and data without rows, or that fails to adapt, would replace the chart and the fullscreen
   * controls with an empty or error state. Returns the notice when the previous data stays.
   */
  function applyChartData(
    response: ChartDataResponse,
    queried: WireFlintInput,
  ): string | undefined {
    const previous = { dataOverride, presentationSpec };
    const { input, ...payload } = response;
    // A refresh returns the effective input, with the limits the server applied (a facet cap,
    // say); without one, the data answers the input as requested.
    dataOverride = { queried: input ?? queried, payload };
    // The refreshed input carries the presentation edits the request was built from.
    presentationSpec = undefined;
    const outcome = view.kind;
    if (outcome === "ready") return undefined;
    dataOverride = previous.dataOverride;
    presentationSpec = previous.presentationSpec;
    return outcome === "empty"
      ? m.mcp_chart_refresh_empty()
      : m.mcp_chart_refresh_failed();
  }

  function applyPresentation(next: WireFlintChartSpec) {
    presentationSpec = next;
    // A pending refresh was built from the previous spec, and its result would replace this edit;
    // it is asked again, with the edit.
    if (refreshing) applyControls(controls);
    updateModelContext();
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
    onControlsChange={applyControls}
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
