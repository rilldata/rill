<script lang="ts">
  import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
  import DashboardStateLoader from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateLoader.svelte";
  import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";

  /**
   * Test component that adds some text to assert dashboard load.
   */
  export let exploreName: string;
  export let bookmarkOrTokenExploreState:
    | CompoundQueryResult<Partial<MetricsExplorerEntity> | undefined>
    | undefined = undefined;
</script>

<DashboardStateLoader {exploreName} {bookmarkOrTokenExploreState}>
  <div>Dashboard loaded!</div>
</DashboardStateLoader>
