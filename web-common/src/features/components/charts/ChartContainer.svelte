<script lang="ts">
  import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
  import { transformChartSpecToPivotState } from "@rilldata/web-common/features/components/charts/explore-transformer";
  import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { useExploreAvailability } from "@rilldata/web-common/features/explore-mappers/explore-validation";
  import { transformTimeAndFiltersToExploreState } from "@rilldata/web-common/features/explores/explore-link/explore-state-transformer";
  import ExploreLink from "@rilldata/web-common/features/explores/explore-link/ExploreLink.svelte";
  import { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
  import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { Color } from "chroma-js";
  import type { Readable } from "svelte/store";
  import { derived, readable } from "svelte/store";
  import { CHART_CONFIG } from "./config";
  import { getChartData } from "./data-provider";
  import type { ChartProvider, ChartSpec, ChartType } from "./types";

  export let chartType: ChartType;
  export let spec: Readable<ChartSpec>;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let theme: "light" | "dark" = "light";
  export let themeStore: Readable<{ primary?: Color; secondary?: Color }> =
    readable({});
  export let showExploreLink: boolean = false;
  export let organization: string | undefined = undefined;
  export let project: string | undefined = undefined;

  let chartProvider: ChartProvider;
  $: console.log(chartType, $spec);
  $: {
    const chartConfig = CHART_CONFIG[chartType];
    chartProvider = new chartConfig.provider(spec, {});
  }

  $: metricsViewSelectors = new MetricsViewSelectors($runtime.instanceId);

  $: measures = metricsViewSelectors.getMeasuresForMetricView(
    $spec.metrics_view,
  );

  $: chartDataQuery = chartProvider.createChartDataQuery(
    runtime,
    timeAndFilterStore,
  );

  $: chartData = getChartData({
    config: $spec,
    chartDataQuery,
    metricsView: metricsViewSelectors,
    themeStore,
    timeAndFilterStore,
    getDomainValues: () => chartProvider.getChartDomainValues($measures),
    isDarkMode: theme === "dark",
  });

  $: exploreAvailability = showExploreLink
    ? useExploreAvailability($runtime.instanceId, $spec?.metrics_view)
    : readable({ isAvailable: false, exploreName: null });

  $: exploreName = derived(
    exploreAvailability,
    (availability) => availability?.exploreName ?? $spec?.metrics_view,
  );

  $: exploreState = derived(
    [timeAndFilterStore, chartProvider.combinedWhere, exploreName],
    ([timeAndFilter, filterState, expName]) => {
      if (!showExploreLink || !expName) return undefined;

      const { dimensionFilters, dimensionThresholdFilters } =
        splitWhereFilter(filterState);
      const baseState = transformTimeAndFiltersToExploreState(timeAndFilter);
      const pivotState = transformChartSpecToPivotState(
        $spec,
        timeAndFilter.timeGrain,
      );

      return {
        ...baseState,
        whereFilter: dimensionFilters,
        dimensionThresholdFilters,
        activePage: DashboardState_ActivePage.PIVOT,
        pivot: pivotState,
        showTimeComparison: false,
      };
    },
  );
</script>

{#if $spec}
  <div class="size-full flex flex-col">
    <div class="flex-1">
      <Chart
        {chartType}
        chartSpec={$spec}
        {chartData}
        measures={$measures}
        {theme}
        isCanvas={true}
      />
    </div>
    {#if showExploreLink && $exploreAvailability.isAvailable}
      <div class="flex justify-end p-2">
        <ExploreLink
          exploreName={$exploreName}
          {organization}
          {project}
          exploreState={$exploreState}
          mode="icon-button"
        />
      </div>
    {/if}
  </div>
{/if}
