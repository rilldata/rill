<script lang="ts">
  import type { OtherSourceOfState } from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateLoader.svelte";
  import DashboardStateLoader from "web-common/src/features/dashboards/state-managers/loaders/DashboardStateLoader.svelte";

  /**
   * Test component that adds some text to assert dashboard load.
   */
  export let exploreName: string;
  export let otherSourcesOfState: OtherSourceOfState[];
</script>

<DashboardStateLoader {exploreName} {otherSourcesOfState}>
  <div>Dashboard loaded!</div>
</DashboardStateLoader>
