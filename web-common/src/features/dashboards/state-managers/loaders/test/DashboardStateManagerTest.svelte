<script lang="ts">
  import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
  import DashboardStateManager from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateManager.svelte";
  import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";

  /**
   * Test component that adds some text to assert dashboard load.
   */
  export let exploreName: string;
  export let bookmarkOrTokenExploreState:
    | CompoundQueryResult<Partial<ExploreState> | null>
    | undefined = undefined;
</script>

<DashboardStateManager {exploreName} {bookmarkOrTokenExploreState}>
  <div>Dashboard loaded!</div>
</DashboardStateManager>
