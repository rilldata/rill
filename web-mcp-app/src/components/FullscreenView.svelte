<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";
  import type { AdaptedChart } from "../lib/adapt";
  import type { WireFlintChartSpec } from "../lib/contract";
  import {
    topNChoices,
    type ControlOptions,
    type ControlState,
  } from "../lib/controls";
  import type { FlintPanel } from "../lib/flint-options";
  import {
    NO_NOTICES,
    truncationBanner,
    type CompileNotices,
  } from "../lib/notices";
  import CardFooter from "./CardFooter.svelte";
  import CardHeader from "./CardHeader.svelte";
  import ChartSurface from "./ChartSurface.svelte";
  import DataTable from "./DataTable.svelte";
  import SegmentedControl from "./SegmentedControl.svelte";
  import TruncatedBanner from "./TruncatedBanner.svelte";

  type Tab = "chart" | "data";

  let {
    chart,
    isDark,
    fontFamily,
    safeAreaInsets,
    controls,
    controlState,
    panel,
    refreshing = false,
    refreshError,
    openUrl,
    onOpenLink,
    onExit,
    onControlsChange,
    onPresentationChange,
  }: {
    chart: AdaptedChart;
    isDark: boolean;
    fontFamily: string;
    safeAreaInsets: McpUiHostContext["safeAreaInsets"];
    /** Undefined when the host cannot run server tools or the chart has nothing to control. */
    controls: ControlOptions | undefined;
    controlState: ControlState;
    /** Flint charts only: the presentation controls, applied client-side. */
    panel?: FlintPanel;
    refreshing?: boolean;
    refreshError?: string;
    openUrl: string | undefined;
    onOpenLink: () => void;
    onExit: () => void;
    onControlsChange: (next: ControlState) => void;
    onPresentationChange?: (next: WireFlintChartSpec) => void;
  } = $props();

  let tab = $state<Tab>("chart");
  let notices = $state.raw<CompileNotices>(NO_NOTICES);
  const banner = $derived(truncationBanner(chart, notices));

  const tabOptions: { value: Tab; label: string }[] = [
    { value: "chart", label: m.mcp_chart_tab_chart() },
    { value: "data", label: m.mcp_chart_tab_data() },
  ];

  const grainOptions = $derived(
    (controls?.grains ?? []).map((grain: V1TimeGrain) => ({
      value: grain,
      label: TIME_GRAIN[grain]?.label ?? grain,
    })),
  );
  const topNOptions = $derived(
    topNChoices(controlState.topN).map((n) => ({
      value: String(n),
      label: String(n),
    })),
  );
  const comparisonOptions = [
    { value: "on", label: m.mcp_chart_comparison_on() },
    { value: "off", label: m.mcp_chart_comparison_off() },
  ];

  const padding = $derived(
    `${safeAreaInsets?.top ?? 0}px ${safeAreaInsets?.right ?? 0}px ${safeAreaInsets?.bottom ?? 0}px ${safeAreaInsets?.left ?? 0}px`,
  );
</script>

<div class="fullscreen" style:padding>
  <div class="top">
    <CardHeader title={chart.title} caption={chart.caption} />
  </div>

  <div class="toolbar">
    <SegmentedControl
      label={m.mcp_chart_view_label()}
      options={tabOptions}
      value={tab}
      onchange={(next) => (tab = next as Tab)}
    />
    {#if controls}
      {#if grainOptions.length > 1}
        <SegmentedControl
          label={m.mcp_chart_control_grain()}
          options={grainOptions}
          value={controlState.grain}
          onchange={(grain) =>
            onControlsChange({ ...controlState, grain: grain as V1TimeGrain })}
        />
      {/if}
      {#if controls.topN}
        <SegmentedControl
          label={m.mcp_chart_control_top_n()}
          options={topNOptions}
          value={controlState.topN === undefined
            ? undefined
            : String(controlState.topN)}
          onchange={(topN) =>
            onControlsChange({ ...controlState, topN: Number(topN) })}
        />
      {/if}
      {#if controls.comparison}
        <SegmentedControl
          label={m.mcp_chart_control_comparison()}
          options={comparisonOptions}
          value={controlState.comparison ? "on" : "off"}
          onchange={(comparison) =>
            onControlsChange({
              ...controlState,
              comparison: comparison === "on",
            })}
        />
      {/if}
    {/if}
  </div>

  {#if panel && tab === "chart"}
    <!-- Presentation controls from Flint's option model: every change recompiles locally. -->
    <div class="toolbar presentation">
      {#if panel.chartType}
        <SegmentedControl
          label={m.mcp_chart_control_chart_type()}
          options={panel.chartType.choices}
          value={panel.chartType.current}
          onchange={(type) =>
            onPresentationChange?.(panel.apply.chartType(type))}
        />
      {/if}
      {#if panel.layout}
        <SegmentedControl
          label={m.mcp_chart_control_layout()}
          options={panel.layout.choices}
          value={panel.layout.value}
          onchange={(id) => onPresentationChange?.(panel.apply.layout(id))}
        />
      {/if}
      {#each panel.properties as property (property.key)}
        <SegmentedControl
          label={property.label}
          options={property.choices}
          value={property.value}
          onchange={(choice) =>
            onPresentationChange?.(panel.apply.property(property.key, choice))}
        />
      {/each}
    </div>
  {/if}

  {#if refreshError}
    <p class="notice" role="status">{refreshError}</p>
  {/if}

  <div class="body">
    {#if tab === "chart"}
      <ChartSurface
        {chart}
        {isDark}
        {fontFamily}
        fill
        dimmed={refreshing}
        onCompiled={(next) => (notices = next)}
      />
    {:else}
      <div class="table">
        <DataTable table={chart.table} />
      </div>
    {/if}
  </div>

  {#if banner}
    <TruncatedBanner note={banner.note} detail={banner.detail} />
  {/if}

  <CardFooter
    {openUrl}
    {onOpenLink}
    secondary="exit-fullscreen"
    onSecondary={onExit}
  />
</div>

<style lang="postcss">
  .fullscreen {
    @apply flex flex-col w-full box-border bg-surface-background;
    height: 100dvh;
  }
  .toolbar {
    @apply flex flex-wrap items-center gap-x-4 gap-y-2 px-4 py-2;
  }
  .toolbar.presentation {
    @apply pt-0;
  }
  .body {
    @apply flex flex-1 min-h-0 flex-col px-2;
  }
  .table {
    @apply flex-1 min-h-0 px-2 pb-2;
  }
  .notice {
    @apply mx-4 my-1 text-xs text-fg-tertiary;
  }
</style>
