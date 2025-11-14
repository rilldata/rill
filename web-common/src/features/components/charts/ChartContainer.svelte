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

  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import FilterChipsReadOnly from "@rilldata/web-common/features/dashboards/filters/FilterChipsReadOnly.svelte";
  import type { Readable } from "svelte/store";
  import { derived, readable } from "svelte/store";
  import { Theme } from "../../themes/theme";
  import { CHART_CONFIG } from "./config";
  import { getChartData } from "./data-provider";
  import type { ChartProvider, ChartSpec, ChartType } from "./types";

  export let chartType: ChartType;
  export let spec: Readable<ChartSpec>;
  export let timeAndFilterStore: Readable<TimeAndFilterStore>;
  export let themeMode: "light" | "dark" = "light";
  /**
   * Full theme object with all CSS variables for current mode
   * If not provided, chart will fall back to defaults
   */
  export let theme: Record<string, string> | undefined = undefined;
  export let themeStore: Readable<Theme> = readable(new Theme(undefined));
  export let showExploreLink: boolean = false;
  export let organization: string | undefined = undefined;
  export let project: string | undefined = undefined;

  let chartProvider: ChartProvider;
  $: {
    const chartConfig = CHART_CONFIG[chartType];
    chartProvider = new chartConfig.provider(spec, {});
  }

  $: metricsViewSelectors = new MetricsViewSelectors($runtime.instanceId);

  $: measures = metricsViewSelectors.getMeasuresForMetricView(
    $spec.metrics_view,
  );

  $: dimensions = metricsViewSelectors.getDimensionsForMetricView(
    $spec.metrics_view,
  );

  $: chartDataQuery = chartProvider.createChartDataQuery(
    runtime,
    timeAndFilterStore,
  );

  $: ({ dimensionFilters: whereFilter, dimensionThresholdFilters } =
    splitWhereFilter($timeAndFilterStore.where));

  $: chartData = getChartData({
    config: $spec,
    chartDataQuery,
    metricsView: metricsViewSelectors,
    themeStore,
    timeAndFilterStore,
    getDomainValues: () => chartProvider.getChartDomainValues($measures),
    isThemeModeDark: themeMode === "dark",
  });

  $: chartTitle = chartProvider?.chartTitle?.($chartData.fields) ?? "";

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
    {#if chartTitle}
      <div class="flex items-center justify-between px-4 py-2">
        <div
          class="flex items-center gap-x-2 w-full max-w-full overflow-x-auto chip-scroll-container"
        >
          <h4 class="title">{chartTitle}</h4>
          {#if "metrics_view" in $spec}
            <Filter size="16px" className="text-gray-400 flex-shrink-0" />
            <FilterChipsReadOnly
              metricsViewNames={[$spec.metrics_view]}
              dimensions={$dimensions}
              measures={$measures}
              {dimensionThresholdFilters}
              dimensionsWithInlistFilter={[]}
              filters={whereFilter}
              displayTimeRange={$timeAndFilterStore.timeRange}
              queryTimeStart={$timeAndFilterStore.timeRange.start}
              queryTimeEnd={$timeAndFilterStore.timeRange.end}
              hasBoldTimeRange={false}
              chipLayout="scroll"
            />
          {/if}
        </div>
        {#if showExploreLink && $exploreAvailability.isAvailable}
          <ExploreLink
            exploreName={$exploreName}
            {organization}
            {project}
            exploreState={$exploreState}
            mode="icon-button"
          />
        {/if}
      </div>
    {/if}
    <div class="flex-1">
      <Chart
        {chartType}
        chartSpec={$spec}
        {chartData}
        measures={$measures}
        {themeMode}
        {theme}
        isCanvas={true}
      />
    </div>
  </div>
{/if}

<style lang="postcss">
  .title {
    font-size: 15px;
    line-height: 26px;
    @apply flex-shrink-0;
    @apply font-medium text-gray-800 truncate;
  }

  .chip-scroll-container {
    mask-image: linear-gradient(to right, black 95%, transparent);
    -webkit-mask-image: linear-gradient(to right, black 95%, transparent);
    mask-size: 100% 100%;
    mask-repeat: no-repeat;
    -webkit-mask-size: 100% 100%;
    -webkit-mask-repeat: no-repeat;
  }
</style>
