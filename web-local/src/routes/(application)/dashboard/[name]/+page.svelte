<script lang="ts">
  import { DashboardWorkspace } from "@rilldata/web-common/features/dashboards";

  export let data;

  $: metricViewName = data.metricViewName;
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

<DashboardWorkspace {metricViewName} />
