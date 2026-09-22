<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import Filters from "@rilldata/web-common/features/dashboards/filters/Filters.svelte";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import { AD_BIDS_METRICS_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";

  /**
   * Test component that renders the filter bar the way the explore dashboard does, so that tests
   * can drive filters through the UI and assert on the explore state and the url they produce.
   *
   * `Tooltip.Provider` stands in for the one the app layout provides, which the time filters need.
   */
  export let exploreName: string;
</script>

<Tooltip.Provider>
  <StateManagersProvider metricsViewName={AD_BIDS_METRICS_NAME} {exploreName}>
    <DashboardStateManager {exploreName}>
      <Filters />
      <div>Dashboard loaded!</div>
    </DashboardStateManager>
  </StateManagersProvider>
</Tooltip.Provider>
