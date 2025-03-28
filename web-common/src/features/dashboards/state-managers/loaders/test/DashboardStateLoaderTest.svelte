<script lang="ts">
  import DashboardStateLoader from "@rilldata/web-common/features/dashboards/state-managers/loaders/DashboardStateLoader.svelte";

  /**
   * Test component that adds some text to assert dashboard load.
   */
  export let exploreName: string;
  export let otherSourcesOfState: any[];
</script>

<DashboardStateLoader {exploreName} {otherSourcesOfState}>
  <div>Dashboard loaded!</div>
</DashboardStateLoader>
