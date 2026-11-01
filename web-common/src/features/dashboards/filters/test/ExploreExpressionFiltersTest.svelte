<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Filters from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { AD_BIDS_METRICS_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";

  /**
   * Test component that renders the filter bar the way the explore dashboard does, so that tests
   * can drive filters through the UI and assert on the explore state and the url they produce.
   */
  export let exploreName: string;
  export let hasTimeSeries: boolean = false;
</script>

<!-- The app layout supplies the tooltip provider, which the advanced filter pill needs. -->
<Tooltip.Provider>
  <StateManagersProvider metricsViewName={AD_BIDS_METRICS_NAME} {exploreName}>
    <DashboardStateManager {exploreName}>
      <Filters
        timeRanges={[]}
        metricsViewName={AD_BIDS_METRICS_NAME}
        {hasTimeSeries}
      />
      <div>Dashboard loaded!</div>
    </DashboardStateManager>
  </StateManagersProvider>
</Tooltip.Provider>
