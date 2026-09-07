<script lang="ts">
  import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
  import { transformChartSpecToPivotState } from "@rilldata/web-common/features/components/charts/explore-transformer";
  import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { useExploreAvailability } from "@rilldata/web-common/features/explore-mappers/explore-validation";
  import type { ExploreAvailabilityResult } from "@rilldata/web-common/features/explore-mappers/types";
  import { transformTimeAndFiltersToExploreState } from "@rilldata/web-common/features/explores/explore-link/explore-state-transformer";
  import ExploreLink from "@rilldata/web-common/features/explores/explore-link/ExploreLink.svelte";
  import { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
  import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import Filter from "@rilldata/web-common/components/icons/Filter.svelte";
  import ThemeProvider from "@rilldata/web-common/features/dashboards/ThemeProvider.svelte";
  import { THEME_STORE_CONTEXT_KEY } from "@rilldata/web-common/features/themes/theme-boundary";
  import { getContext, hasContext, onDestroy } from "svelte";
  import type { View } from "svelte-vega";
  import type { Readable, Writable } from "svelte/store";
  import { readable } from "svelte/store";
  import { Theme } from "../../themes/theme";
  import { CHART_CONFIG } from "./config";
  import { getChartData } from "./data-provider";
  import type { ChartSpec, ChartType } from "./types";
  import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
  import ReadonlyExpressionFilters from "@rilldata/web-common/features/dashboards/filters/ReadonlyExpressionFilters.svelte";
  import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
  import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

  const hasParentTheme = hasContext(THEME_STORE_CONTEXT_KEY);
  const parentThemeStore = getContext<Writable<Theme | undefined>>(
    THEME_STORE_CONTEXT_KEY,
  );

  let {
    chartType,
    spec,
    timeAndFilterStore,
    themeMode = "light",
    theme = new Theme(undefined),
    showExploreLink = false,
    organization = undefined,
    project = undefined,
  }: {
    chartType: ChartType;
    spec: Readable<ChartSpec>;
    timeAndFilterStore: Readable<TimeAndFilterStore>;
    themeMode?: "light" | "dark";
    theme?: Theme;
    showExploreLink?: boolean;
    organization?: string | undefined;
    project?: string | undefined;
  } = $props();

  const client = useRuntimeClient();
  let view = $state() as View;

  let chartProvider = $derived.by(() => {
    const chartConfig = CHART_CONFIG[chartType];
    return new chartConfig.provider(spec, {});
  });

  // Use parent theme from context if available, otherwise use the prop
  let effectiveTheme = $derived(hasParentTheme ? $parentThemeStore : theme);

  /**
   * Full theme object with all CSS variables for current mode
   * If not provided, chart will fall back to defaults
   */
  let currentTheme = $derived(
    effectiveTheme?.resolvedThemeObject?.[
      themeMode === "dark" ? "dark" : "light"
    ],
  );

  const metricsViewProvider = new MetricsViewsProvider(client, []);
  $effect(() =>
    metricsViewProvider.setMetricsViewNames(
      $spec.metrics_view ? [$spec.metrics_view] : [],
    ),
  );
  const yamlConfigProvider = new YAMLConfigProvider();
  const expressionFilterManager = new ExpressionFilterManager(
    metricsViewProvider,
    yamlConfigProvider,
  );
  $effect(() =>
    expressionFilterManager.setExprForMetricsView(
      $spec.metrics_view,
      $timeAndFilterStore.where,
    ),
  );

  let metricsViewSelectors = $derived(new MetricsViewSelectors(client));

  let chartDataQuery = $derived(
    chartProvider.createChartDataQuery(client, timeAndFilterStore),
  );

  let chartData = $derived(
    getChartData({
      config: $spec,
      chartDataQuery,
      metricsView: metricsViewSelectors,
      themeStore: readable(effectiveTheme),
      timeAndFilterStore,
      getDomainValues: () =>
        chartProvider.getChartDomainValues(metricsViewProvider.measures),
      isThemeModeDark: themeMode === "dark",
    }),
  );

  let chartTitle = $derived(
    chartProvider?.chartTitle?.($chartData.fields) ?? "",
  );

  let exploreAvailabilityStore = $derived(
    showExploreLink
      ? useExploreAvailability(client, $spec?.metrics_view)
      : readable<ExploreAvailabilityResult>({ isAvailable: false }),
  );
  let exploreAvailability = $derived($exploreAvailabilityStore);

  let exploreName = $derived(
    exploreAvailability?.exploreName ?? $spec?.metrics_view,
  );

  let combinedWhere = $derived(chartProvider.combinedWhere);

  let exploreState = $derived.by(() => {
    if (!showExploreLink || !exploreName) return undefined;

    const baseState =
      transformTimeAndFiltersToExploreState($timeAndFilterStore);
    const pivotState = transformChartSpecToPivotState(
      $spec,
      $timeAndFilterStore.timeGrain,
    );

    return {
      ...baseState,
      whereFilter: $combinedWhere,
      activePage: DashboardState_ActivePage.PIVOT,
      pivot: pivotState,
      showTimeComparison: false,
    };
  });

  onDestroy(() => {
    metricsViewProvider.cleanup();
    yamlConfigProvider.cleanup?.();
  });
</script>

{#if $spec}
  <ThemeProvider theme={effectiveTheme}>
    <div class="size-full flex flex-col">
      {#if chartTitle}
        <div class="flex items-center justify-between px-4 py-2">
          <div
            class="flex items-center gap-x-2 w-full max-w-full overflow-x-auto chip-scroll-container"
          >
            <h4 class="title">{chartTitle}</h4>
            {#if "metrics_view" in $spec}
              <Filter size="16px" className="text-fg-secondary flex-shrink-0" />
              <ReadonlyExpressionFilters
                {expressionFilterManager}
                displayTimeRange={$timeAndFilterStore.timeRange}
                displayComparisonTimeRange={$timeAndFilterStore.showTimeComparison
                  ? $timeAndFilterStore.comparisonTimeRange
                  : undefined}
                queryTimeStart={$timeAndFilterStore.timeRange.start}
                queryTimeEnd={$timeAndFilterStore.timeRange.end}
                hasBoldTimeRange={false}
                chipLayout="scroll"
              />
            {/if}
          </div>
          {#if showExploreLink && exploreAvailability.isAvailable}
            <ExploreLink
              {exploreName}
              displayName={exploreAvailability.displayName}
              {organization}
              {project}
              {exploreState}
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
          measures={metricsViewProvider.measures}
          {themeMode}
          theme={currentTheme}
          isCanvas={true}
          bind:view
        />
      </div>
    </div>
  </ThemeProvider>
{/if}

<style lang="postcss">
  .title {
    font-size: 15px;
    line-height: 26px;
    @apply flex-shrink-0;
    @apply font-medium text-fg-primary truncate;
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
