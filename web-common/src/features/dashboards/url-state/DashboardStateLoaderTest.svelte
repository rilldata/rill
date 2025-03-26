<script lang="ts">
  import DashboardStateLoader from "./DashboardStateLoader.svelte";

  export let exploreName: string;
</script>

<DashboardStateLoader {exploreName}>
  <div>Dashboard loaded!</div>
</DashboardStateLoader>
