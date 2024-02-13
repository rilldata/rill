<script lang="ts">
  import { page } from "$app/stores";
  import { ConnectedPreviewTable } from "@rilldata/web-common/components/preview-table";
  import ExternalTableWorkspaceHeader from "@rilldata/web-common/features/external-tables/ExternalTableWorkspaceHeader.svelte";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
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

<WorkspaceContainer assetID={fullyQualifiedTableName} inspector={false}>
  <ExternalTableWorkspaceHeader {fullyQualifiedTableName} slot="header" />
  <ConnectedPreviewTable objectName={table} loading={false} slot="body" />
</WorkspaceContainer>
