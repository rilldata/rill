<script lang="ts">
  import CellInspector from "@rilldata/web-common/components/CellInspector.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { belowSm } from "@rilldata/web-common/lib/store-utils/media-query-store";
  import {
    extractErrorStatusCode,
    isNotFoundError,
  } from "@rilldata/web-common/lib/errors";
  import EphemeralMeasureDialog from "@rilldata/web-common/features/dashboards/ephemeral-measures/EphemeralMeasureDialog.svelte";
  import { ephemeralMeasureDialog } from "@rilldata/web-common/features/dashboards/ephemeral-measures/dialog-store";
  import PivotDisplay from "@rilldata/web-common/features/dashboards/pivot/PivotDisplay.svelte";
  import TabBar from "@rilldata/web-common/features/dashboards/tab-bar/TabBar.svelte";
  import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { navigationOpen } from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { githubStarNudge } from "@rilldata/web-common/features/github-star/github-star.svelte";
  import { onDestroy, onMount } from "svelte";
  import { get, readable, type Readable } from "svelte/store";
  import { useExploreState } from "web-common/src/features/dashboards/stores/dashboard-stores";
  import { DashboardState_ActivePage } from "../../../proto/gen/rill/ui/v1/dashboard_pb";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import { activeDashboardTheme } from "../../themes/active-dashboard-theme";
  import { createResolvedThemeStore } from "../../themes/selectors";
  import MeasuresContainer from "../big-number/MeasuresContainer.svelte";
  import DimensionDisplay from "../dimension-table/DimensionDisplay.svelte";
  import Filters from "../filters/Filters.svelte";
  import { selectedMockUserStore } from "../granular-access-policies/stores";
  import LeaderboardDisplay from "../leaderboard/LeaderboardDisplay.svelte";
  import RowsViewerAccordion from "../rows-viewer/RowsViewerAccordion.svelte";
  import { getStateManagers } from "../state-managers/state-managers";
  import ThemeProvider from "../ThemeProvider.svelte";
  import { useTimeControlStore } from "../time-controls/time-control-store";
  import TimeDimensionDisplay from "../time-dimension-details/TimeDimensionDisplay.svelte";
  import MetricsTimeSeriesCharts from "../time-series/MetricsTimeSeriesCharts.svelte";
  import {
    DEFAULT_TDD_CHART_HEIGHT,
    DEFAULT_TIMESERIES_WIDTH,
    MIN_TDD_CHART_HEIGHT,
    MIN_TIMESERIES_WIDTH,
    exploreTimeseriesWidth,
    tddChartHeight,
  } from "./dashboard-layout-store";

  export let exploreName: string;
  export let metricsViewName: string;
  export let isEmbedded: boolean = false;
  // Stacks the dashboard below `sm` and hands the pivot a "use desktop" notice there.
  // Width alone can't tell a phone from a narrow embed iframe or Rill Developer window,
  // so only Rill Cloud's own explore route opts in.
  export let phoneLayout: boolean = false;
  export let embedThemeName: Readable<string | null> | null = null;

  // Vertical space reserved below the chart for the chart toolbar, axis, big
  // number, and a minimum detail table height when computing the drag maximum.
  const TDD_RESERVED_HEIGHT = 280;
  const StateManagers = getStateManagers();
  const {
    selectors: {
      measures: { visibleMeasures },
      dimensions: { getDimensionByName },
      pivot: { showPivot },
    },
    dashboardStore,
    expressionFilterManager,
  } = StateManagers;

  const { adminServer, cloudDataViewer, readOnly } = featureFlags;

  const timeControlsStore = useTimeControlStore(StateManagers);

  let exploreContainerWidth: number;
  let exploreContainerHeight: number;
  let resizing = false;

  const client = useRuntimeClient();

  $: ({ selectedTimeDimension } = $dashboardStore);
  const filterStore =
    expressionFilterManager.getExprStoreForMetricsView(metricsViewName);
  $: dimensionOnlyFilter = $filterStore?.dimensionOnlyExpr;
  $: whereFilter = $filterStore?.expr;

  $: extraLeftPadding = !$navigationOpen;

  $: exploreState = useExploreState(exploreName);

  $: activePage = $exploreState?.activePage;
  $: showTimeDimensionDetail = Boolean(
    activePage === DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
  );
  $: showDimensionTable = Boolean(
    activePage === DashboardState_ActivePage.DIMENSION_TABLE,
  );

  $: selectedDimensionName = $exploreState?.selectedDimensionName;
  $: selectedDimension =
    selectedDimensionName && $getDimensionByName(selectedDimensionName);
  $: expandedMeasureName = $exploreState?.tdd?.expandedMeasureName;

  $: isRillDeveloper = $readOnly === false;

  // Check if the mock user (if selected) has access to the explore
  $: exploreQuery = useExploreValidSpec(client, exploreName);

  $: ({ data, error: exploreError } = $exploreQuery);

  $: exploreSpec = data?.explore;

  $: hasTimeSeries = !!data?.metricsView?.timeDimension;

  $: mockUserHasNoAccess =
    $selectedMockUserStore && isNotFoundError(exploreError);

  $: hidePivot = isEmbedded && exploreSpec?.embedsHidePivot;

  $: ({
    timeStart: start,
    timeEnd: end,
    showTimeComparison,
    comparisonTimeStart,
    comparisonTimeEnd,
    ready: timeControlsReady = false,
  } = $timeControlsStore);

  $: timeRange = {
    start,
    end,
    timeDimension: selectedTimeDimension,
  };

  $: comparisonTimeRange = showTimeComparison
    ? {
        start: comparisonTimeStart,
        end: comparisonTimeEnd,
        timeDimension: selectedTimeDimension,
      }
    : undefined;

  $: timeRanges = exploreSpec?.timeRanges ?? [];

  $: visibleMeasureNames = $visibleMeasures.map(({ name }) => name ?? "");

  // For non-embedded dashboards, theme can come from URL params.
  // For embedded dashboards, embedThemeName prop takes precedence.
  const urlThemeName = readable<string | null>(null, (set) => {
    set(null);
    return () => {};
  });

  let themeSource: Readable<string | null> = urlThemeName;
  $: themeSource = isEmbedded && embedThemeName ? embedThemeName : urlThemeName;

  $: theme = createResolvedThemeStore(themeSource, exploreQuery, client);

  // Publish the resolved theme to the shared store for external components (e.g., chat in layout)
  $: activeDashboardTheme.set($theme);

  onMount(() => {
    // Github star nudge is Rill developer only.
    // Nudge on dashboard render.
    if (!isEmbedded && !get(adminServer)) githubStarNudge.armPayoff();
  });

  // Clear the active theme when this dashboard is destroyed
  onDestroy(() => activeDashboardTheme.set(undefined));
</script>

<ThemeProvider theme={$theme}>
  <article
    class="flex flex-col overflow-y-hidden bg-surface-background"
    class:phone-layout={phoneLayout}
    bind:clientWidth={exploreContainerWidth}
    class:w-full={$dynamicHeight}
    class:size-full={!$dynamicHeight}
  >
    <div
      id="header"
      class="border-b {phoneLayout
        ? 'w-full sm:w-fit'
        : 'w-fit'} min-w-full flex flex-col bg-surface-subtle slide"
      class:left-shift={extraLeftPadding}
    >
      {#if mockUserHasNoAccess}
        <div class="mb-3"></div>
      {:else}
        {#key exploreName}
          <!-- In the phone layout the tab bar leaves the corner overlay (it would sit on
               top of wrapped filter chips) and flows below the filters. -->
          <section
            class="flex relative justify-between gap-x-4 py-4 px-4 {phoneLayout
              ? 'flex-col sm:flex-row gap-y-2 pb-2 sm:pb-6'
              : 'pb-6'}"
          >
            <Filters {timeRanges} {metricsViewName} {hasTimeSeries} />
            <div
              class="flex flex-col {phoneLayout
                ? 'self-end sm:absolute sm:bottom-0 sm:right-0'
                : 'absolute bottom-0 right-0'}"
            >
              <TabBar {hidePivot} {exploreName} onPivot={$showPivot} />
            </div>
          </section>
        {/key}
      {/if}
    </div>

    {#if mockUserHasNoAccess}
      <!-- Additional safeguard for mock users without dashboard access. -->
      <ErrorPage
        statusCode={extractErrorStatusCode(exploreError)}
        header="This user can't access this dashboard"
        body="The security policy for this dashboard may make contents invisible to you. If you deploy this dashboard, {$selectedMockUserStore?.email} will see a 404."
      />
    {:else if $showPivot}
      {#if phoneLayout && $belowSm}
        <!-- The pivot's table and config sidebar don't fit phones, so a notice takes its place
             and the tab bar above leads back to Explore. `{#if}` keeps the pivot and its queries
             from mounting behind the notice. -->
        <div class="flex flex-1 items-center justify-center p-8">
          <CtaContentContainer>
            <CtaHeader>{m.pivot_desktop_only_title()}</CtaHeader>
            <CtaMessage>{m.pivot_desktop_only_message()}</CtaMessage>
          </CtaContentContainer>
        </div>
      {:else}
        <PivotDisplay {isEmbedded} />
      {/if}
    {:else}
      <div
        class="flex gap-x-1 overflow-hidden slide pb-0 {showTimeDimensionDetail
          ? 'flex-col gap-y-2'
          : phoneLayout
            ? 'flex-col sm:flex-row'
            : 'flex-row'}"
        class:left-shift={extraLeftPadding}
        class:w-full={$dynamicHeight}
        class:size-full={!$dynamicHeight}
        bind:clientHeight={exploreContainerHeight}
      >
        <div
          class="flex-none pl-4 {phoneLayout && !showTimeDimensionDetail
            ? 'h-[50vh] overflow-y-auto sm:h-auto sm:overflow-y-visible'
            : ''}"
          class:max-w-full={phoneLayout}
          class:pt-2={!showTimeDimensionDetail}
          style:width={showTimeDimensionDetail
            ? "auto"
            : `${$exploreTimeseriesWidth}px`}
        >
          {#key exploreName}
            {#if hasTimeSeries}
              <MetricsTimeSeriesCharts
                {exploreName}
                {dimensionOnlyFilter}
                {whereFilter}
                hideStartPivotButton={hidePivot}
                tddChartHeight={$tddChartHeight}
              />
            {:else}
              <MeasuresContainer {metricsViewName} {whereFilter} />
            {/if}
          {/key}
        </div>

        {#if showTimeDimensionDetail && expandedMeasureName}
          <div class="relative flex-none bg-border h-[1px]">
            <Resizer
              direction="NS"
              side="bottom"
              dimension={$tddChartHeight}
              min={MIN_TDD_CHART_HEIGHT}
              max={Math.max(
                MIN_TDD_CHART_HEIGHT,
                (exploreContainerHeight || 600) - TDD_RESERVED_HEIGHT,
              )}
              basis={DEFAULT_TDD_CHART_HEIGHT}
              bind:resizing
              onUpdate={(height: number) => {
                tddChartHeight.set(height);
              }}
            />
          </div>
          <TimeDimensionDisplay
            {exploreName}
            {expandedMeasureName}
            hideStartPivotButton={hidePivot}
          />
        {:else}
          <div
            class="relative flex-none bg-border w-[1px] {phoneLayout
              ? 'hidden sm:block'
              : ''}"
          >
            <Resizer
              dimension={$exploreTimeseriesWidth}
              min={MIN_TIMESERIES_WIDTH}
              max={exploreContainerWidth - 500}
              basis={DEFAULT_TIMESERIES_WIDTH}
              bind:resizing
              side="right"
              onUpdate={(width: number) => {
                exploreTimeseriesWidth.set(width);
              }}
            />
          </div>
          <div
            class="pt-2 pl-1 overflow-auto w-full"
            class:min-h-0={phoneLayout}
          >
            {#if showDimensionTable && selectedDimension}
              <DimensionDisplay
                dimension={selectedDimension}
                {metricsViewName}
                {whereFilter}
                {timeRange}
                {comparisonTimeRange}
                {timeControlsReady}
                {visibleMeasureNames}
                hideStartPivotButton={hidePivot}
              />
            {:else}
              <LeaderboardDisplay
                {metricsViewName}
                {whereFilter}
                {timeRange}
                {comparisonTimeRange}
                {timeControlsReady}
              />
            {/if}
          </div>
        {/if}
      </div>
    {/if}

    <CellInspector />

    {#if (isRillDeveloper || $cloudDataViewer) && !showTimeDimensionDetail && !mockUserHasNoAccess}
      <RowsViewerAccordion {metricsViewName} {exploreName} {whereFilter} />
    {/if}
  </article>

  {#if $ephemeralMeasureDialog}
    <EphemeralMeasureDialog />
  {/if}
</ThemeProvider>

<style lang="postcss">
  .left-shift {
    @apply pl-8;
  }

  /* Clears the floating nav-toggle button; phones have no room to spare
     for the indent and the toggle overlays content anyway. */
  .phone-layout .left-shift {
    @apply pl-0 sm:pl-8;
  }
</style>
