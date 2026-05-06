<script lang="ts">
  import { page } from "$app/stores";
  import TablePreviewWorkspace from "@rilldata/web-common/features/connectors/olap/TablePreviewWorkspace.svelte";
  import { readOnly } from "@rilldata/web-common/features/app-flags";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  $: name = $page.params.name;
  $: table = $page.params.table;

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });
</script>

<svelte:head>
  <title>Rill Developer | {table}</title>
</svelte:head>

<TablePreviewWorkspace connector={name} {table} />
