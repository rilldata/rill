<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import SimpleMessage from "../../layout/inspector/SimpleMessage.svelte";
  import { createConnectorServiceOLAPGetTable } from "../../runtime-client";
  import TableInspector from "../connectors/olap/TableInspector.svelte";
  import ReconcilingSpinner from "../entity-management/ReconcilingSpinner.svelte";
  import { fileArtifacts } from "../entity-management/file-artifacts";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";

  const queryClient = useQueryClient();

  export let filePath: string;
  export let connector: string;
  export let database: string;
  export let databaseSchema: string;
  export let table: string;

  $: ({ instanceId } = $runtime);

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: ({ remoteContent } = fileArtifact);
  $: parseError = fileArtifact.getParseError(queryClient, instanceId);
  $: resource = fileArtifact.getResource(queryClient, instanceId);
  $: ({
    isLoading: isResourceLoading,
    error: resourceError,
    data: resourceData,
  } = $resource);
  $: resourceReconcileError = resourceData?.meta?.reconcileError;

  $: tableQuery = createConnectorServiceOLAPGetTable(
    {
      instanceId,
      connector,
      database,
      databaseSchema,
      table,
    },
    {
      query: {
        enabled: !!connector && !!table,
      },
    },
  );
  $: ({ error: tableError, isLoading: isTableLoading } = $tableQuery);
</script>

<Inspector {filePath}>
  {#if !$remoteContent}
    <SimpleMessage
      message={`For help building dashboards, see:<br /><a
        href="https://docs.rilldata.com/build/dashboards"
        target="_blank"
        rel="noopener noreferrer">https://docs.rilldata.com/build/dashboards</a>`}
      includesHtml
    />
  {:else if $parseError}
    <!-- The editor will show actual validation errors -->
    <SimpleMessage message="Fix the errors in the file to continue." />
  {:else if isResourceLoading}
    <div class="spinner-wrapper">
      <ReconcilingSpinner />
    </div>
  {:else if resourceError}
    <SimpleMessage message="Error: {resourceError?.response?.data?.message}" />
  {:else if resourceReconcileError}
    <!-- The editor will show actual validation errors -->
    <SimpleMessage message="Fix the errors in the file to continue." />
  {:else if isTableLoading}
    <div class="spinner-wrapper">
      <ReconcilingSpinner />
    </div>
  {:else if tableError}
    <SimpleMessage message="Error: {tableError?.response?.data?.message}" />
  {:else}
    <TableInspector {connector} {database} {databaseSchema} {table} />
  {/if}
</Inspector>

<style lang="postcss">
  .spinner-wrapper {
    @apply px-4 py-8 size-full;
    @apply flex items-center justify-center;
  }
</style>
