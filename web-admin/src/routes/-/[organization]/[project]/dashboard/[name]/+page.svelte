<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";

  const metricViewName: string = $page.params.name;
  $: console.log("metricViewName", metricViewName);

  // TODO: add 404 logic as in `web-local`'s `dashboard/[name]/+page.svelte`
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

<div class="p-2">
  <Dashboard {metricViewName} />
</div>
