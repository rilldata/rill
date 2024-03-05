<script lang="ts">
  import { page } from "$app/stores";

  $: chartName = $page.params.name;
</script>

<svelte:head>
  <title>Rill Developer | {chartName}</title>
</svelte:head>

TODO
