<script lang="ts">
  import type { OtherSourceOfState } from "@rilldata/web-common/features/dashboards/url-state/DashboardStateLoader.svelte";
  import DashboardStateLoader from "./DashboardStateLoader.svelte";

  export let exploreName: string;
  export let otherSourcesOfState: OtherSourceOfState[];
</script>

<DashboardStateLoader {exploreName} {otherSourcesOfState}>
  <div>Dashboard loaded!</div>
</DashboardStateLoader>
