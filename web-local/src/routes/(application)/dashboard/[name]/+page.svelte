<script lang="ts">
  import { DashboardWorkspace } from "@rilldata/web-local/lib/components/workspace";
  export let data;

  $: metricViewName = data.metricViewName;
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

<DashboardWorkspace {metricViewName} />
