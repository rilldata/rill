<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Callout } from "@rilldata/web-common/components/callout";
  import { humanReadableErrorMessage } from "@rilldata/web-common/features/sources/add-source/errors.js";
  import { refreshSource } from "@rilldata/web-common/features/sources/refreshSource";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServiceListConnectors,
    createRuntimeServicePutFileAndReconcile,
    createRuntimeServiceRefreshAndReconcile,
    getRuntimeServiceGetCatalogEntryQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { parseDocument } from "yaml";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import { fileArtifactsStore } from "../../entity-management/file-artifacts-store";
  import { EntityType } from "../../entity-management/types";

  export let sourceName: string;

  let sourcePath: string;
  $: sourcePath = getFilePathFromNameAndType(sourceName, EntityType.Table);
  $: getSource = createRuntimeServiceGetFile($runtime?.instanceId, sourcePath);
  $: source = parseDocument($getSource?.data?.blob || "{}").toJS();

  $: connectors = createRuntimeServiceListConnectors();

  // get the connector for this source type, if valid
  $: currentConnector = $connectors?.data?.connectors?.find(
    (connector) => connector?.name === source?.type
  );
  $: allConnectors = $connectors?.data?.connectors?.map(
    (connector) => connector.name
  );
  $: remoteConnectorNames = allConnectors?.filter(
    (name) => name !== "local_file"
  );

  const queryClient = useQueryClient();
  const createSource = createRuntimeServicePutFileAndReconcile();
  const refreshSourceMutation = createRuntimeServiceRefreshAndReconcile();

  const onRefreshClick = async (tableName: string) => {
    try {
      await refreshSource(
        currentConnector?.name,
        tableName,
        $runtime?.instanceId,
        $refreshSourceMutation,
        $createSource,
        queryClient
      );
      const queryKey = getRuntimeServiceGetCatalogEntryQueryKey(
        $runtime?.instanceId,
        tableName
      );
      await queryClient.refetchQueries(queryKey);
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };

  $: isReconciling = $fileArtifactsStore.entities[sourcePath]?.isReconciling;
  let uploadErrors = undefined;
  $: uploadErrors = $fileArtifactsStore.entities[sourcePath]?.errors;
</script>

<div
  class="errors flex flex-col items-center pt-8 gap-y-4 m-auto mt-0 text-gray-500"
  style:width="500px"
>
  {#if !allConnectors?.includes(source?.type)}
    <div>
      {#if source?.type}
        Rill does not support a connector for <span class="font-bold"
          >{source?.type}</span
        >.
      {:else}
        Connector not defined.
      {/if}
      Edit <b>{`sources/${sourceName}.yaml`}</b> to add a valid "type:
      {"<filetype>"}" to get started.
    </div>
    <div>
      For more information,
      <a href="https://docs.rilldata.com/develop/import-data"
        >view the documentation</a
      >.
    </div>
  {:else if source?.type === "local_file"}
    <div class="text-center">
      The data file for <span class="font-bold">{sourceName}</span> has not been
      imported as a source.
    </div>
    <Button
      type="primary"
      on:click={async () => {
        uploadErrors = undefined;
        await onRefreshClick(sourceName);
      }}
      >Import a CSV or Parquet file
    </Button>
  {:else if !Object.keys(source || {})?.length}
    <!-- source is empty -->
    <div>
      The source <span class="font-bold">{sourceName}</span>
      is empty. Edit <b>{`sources/${sourceName}.yaml`}</b> to add a source definition.
    </div>
    <div>
      For more information,
      <a href="https://docs.rilldata.com/develop/import-data"
        >view the documentation</a
      >.
    </div>
  {:else if !source?.type}
    <div>
      The source <span class="font-bold">{sourceName}</span> does not have a
      defined type. Edit <b>{`sources/${sourceName}.yaml`}</b> to add "type:
      {"<filetype>"}"
    </div>
    <div>
      For more information,
      <a href="https://docs.rilldata.com/develop/import-data"
        >view the documentation</a
      >.
    </div>
  {:else if remoteConnectorNames.includes(currentConnector?.name) && !source?.uri}
    <div>
      The source URI has not been defined. Edit <b
        >{`sources/${sourceName}.yaml`}</b
      >
      to add "uri:
      {"<uri>"}"
    </div>
    <div>
      For more information,
      <a href="https://docs.rilldata.com/develop/import-data"
        >view the documentation</a
      >.
    </div>
    <!-- {:else if isReconciling}
            <div class="text-center">
              The source <span class="font-bold">{sourceName}</span> is being imported.
            </div>
            <div class="text-center">This may take a few seconds.</div> -->
  {:else}
    <div class="text-center">
      The source <span class="font-bold">{sourceName}</span> has not been imported.
    </div>
  {/if}
  <!-- show any remaining errors -->
  {#if uploadErrors?.length}
    <Callout level="error">
      {#each uploadErrors as error}
        {humanReadableErrorMessage(currentConnector?.name, 3, error.message)}
      {/each}
    </Callout>
  {/if}
</div>

<style>
  .errors > div:not(.text-center) {
    text-align: left;
    width: 500px;
  }
</style>
