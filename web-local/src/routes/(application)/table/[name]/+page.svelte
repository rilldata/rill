<script lang="ts">
  import { page } from "$app/stores";
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import TableWorkspaceHeader from "@rilldata/web-common/features/tables/TableWorkspaceHeader.svelte";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";

  const { readOnly } = featureFlags;

  $: fullyQualifiedTableName = $page.params.name; // `database.table`
  $: [, table] = fullyQualifiedTableName.split(".");

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });
</script>

<svelte:head>
  <title>Rill Developer | {fullyQualifiedTableName}</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <TableWorkspaceHeader {fullyQualifiedTableName} slot="header" />
  <ConnectedPreviewTable objectName={table} loading={false} slot="body" />
</WorkspaceContainer>
