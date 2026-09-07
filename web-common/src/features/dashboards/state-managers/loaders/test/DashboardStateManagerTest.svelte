<script lang="ts">
  import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import StateManagersProvider from "@rilldata/web-common/features/dashboards/state-managers/StateManagersProvider.svelte";
  import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
  import { AD_BIDS_METRICS_NAME } from "@rilldata/web-common/features/dashboards/stores/test-data/data";

  /**
   * Test component that adds some text to assert dashboard load.
   * DashboardStateManager reads the state managers from context, so it is nested under
   * StateManagersProvider the same way the app nests it.
   */
  export let exploreName: string;
  export let bookmarkOrTokenExploreState:
    | CompoundQueryResult<Partial<ExploreState> | null>
    | undefined = undefined;
  export let disableMostRecentDashboardState: boolean = false;
  export let disableInitSessionDashboardState: boolean = false;
</script>

<StateManagersProvider metricsViewName={AD_BIDS_METRICS_NAME} {exploreName}>
  <DashboardStateManager
    {exploreName}
    {bookmarkOrTokenExploreState}
    {disableMostRecentDashboardState}
    {disableInitSessionDashboardState}
  >
    <div>Dashboard loaded!</div>
  </DashboardStateManager>
</StateManagersProvider>
