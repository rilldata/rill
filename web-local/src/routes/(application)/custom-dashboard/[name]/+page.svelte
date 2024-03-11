<script lang="ts">
  import { page } from "$app/stores";

  $: customDashboardName = $page.params.name;
</script>

<svelte:head>
  <title>Rill Developer | {customDashboardName}</title>
</svelte:head>

TODO
