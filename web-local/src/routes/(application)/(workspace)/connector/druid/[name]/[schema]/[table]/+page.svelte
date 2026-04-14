<script lang="ts">
  import { page } from "$app/stores";
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
  import TablePreviewWorkspace from "@rilldata/web-common/features/connectors/olap/TablePreviewWorkspace.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  const { readOnly } = featureFlags;

  $: name = $page.params.name;
  // Druid does not have a "database" concept
  $: databaseSchema = $page.params.schema;
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

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <TablePreviewWorkspace connector={name} {databaseSchema} {table} />
  </div>
  <DeveloperChat />
</div>
