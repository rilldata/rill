<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createConnectorServiceOLAPGetTable } from "../../runtime-client";
  import TableInspector from "../connectors/olap/TableInspector.svelte";
  import ReconcilingSpinner from "../entity-management/ReconcilingSpinner.svelte";
  import { fileArtifacts } from "../entity-management/file-artifacts";

  const queryClient = useQueryClient();

  export let filePath: string;

  $: resource = fileArtifacts
    .getFileArtifact(filePath)
    .getResource(queryClient, $runtime.instanceId);

  $: connector = $resource.data?.metricsView?.spec?.connector;
  $: database = $resource.data?.metricsView?.spec?.database ?? "";
  $: databaseSchema = $resource.data?.metricsView?.spec?.databaseSchema ?? "";
  $: table = $resource.data?.metricsView?.spec?.table;

  $: tableQuery = createConnectorServiceOLAPGetTable(
    {
      instanceId: $runtime.instanceId,
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
  $: ({ error, isLoading } = $tableQuery);

  function customErrorHandler(errorMessage: string) {
    if (errorMessage === "driver: not found") {
      return "Table not found. Please check the table name and try again.";
    }
    return errorMessage;
  }
</script>

{#if !connector || !table}
  <div class="custom-instructions-wrapper" style:text-wrap="balance">
    <div>
      <p>Table not defined.</p>
      <p>
        Specify a table with <code>table: TABLE_NAME</code>
      </p>
    </div>
  </div>
{:else if isLoading}
  <div class="spinner-wrapper">
    <ReconcilingSpinner />
  </div>
{:else if error}
  <div class="custom-instructions-wrapper" style:text-wrap="balance">
    <p>Error: {customErrorHandler(error.response.data.message)}</p>
  </div>
{:else}
  <TableInspector {connector} {database} {databaseSchema} {table} />
{/if}

<style lang="postcss">
  .custom-instructions-wrapper {
    @apply px-4 py-24 w-full;
    @apply italic text-gray-500 text-center;
  }

  .spinner-wrapper {
    @apply px-4 py-8 size-full;
    @apply flex items-center justify-center;
  }
</style>
