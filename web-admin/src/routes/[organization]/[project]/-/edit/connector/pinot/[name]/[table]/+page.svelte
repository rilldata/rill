<script lang="ts">
  import { page } from "$app/stores";
  import TablePreviewWorkspace from "@rilldata/web-common/features/connectors/olap/TablePreviewWorkspace.svelte";

  $: name = $page.params.name;
  $: table = $page.params.table;
</script>

<svelte:head>
  <title>Rill | {table}</title>
</svelte:head>

<TablePreviewWorkspace connector={name} {table} />
