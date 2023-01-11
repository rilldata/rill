<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { Callout } from "@rilldata/web-common/components/callout";
  import { humanReadableErrorMessage } from "@rilldata/web-common/features/sources/add-source/errors.js";
  import { refreshSource } from "@rilldata/web-common/features/sources/refreshSource";
  import { EntityType } from "@rilldata/web-common/lib/entity";
  import {
    getRuntimeServiceGetCatalogEntryQueryKey,
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
    useRuntimeServiceListConnectors,
    useRuntimeServicePutFileAndReconcile,
    useRuntimeServiceRefreshAndReconcile,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { fileArtifactsStore } from "@rilldata/web-local/lib/application-state-stores/file-artifacts-store";
  import { overlay } from "@rilldata/web-local/lib/application-state-stores/overlay-store";
  import { ConnectedPreviewTable } from "@rilldata/web-local/lib/components/preview-table/index.js";
  import { WorkspaceContainer } from "@rilldata/web-local/lib/components/workspace/index.js";
  import { getFilePathFromNameAndType } from "@rilldata/web-local/lib/util/entity-mappers";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { parseDocument } from "yaml";
  import SourceInspector from "./SourceInspector.svelte";
  import SourceWorkspaceHeader from "./SourceWorkspaceHeader.svelte";

  export let sourceName: string;

  const switchToSource = async (name: string) => {
    if (!name) return;

    appStore.setActiveEntity(name, EntityType.Table);
  };

  $: switchToSource(sourceName);

  $: checkForSourceInCatalog = useRuntimeServiceGetCatalogEntry(
    $runtimeStore?.instanceId,
    sourceName
  );

  let sourcePath: string;
  $: sourcePath = getFilePathFromNameAndType(sourceName, EntityType.Table);
  $: getSource = useRuntimeServiceGetFile(
    $runtimeStore?.instanceId,
    sourcePath
  );

  $: source = parseDocument($getSource?.data?.blob || "{}").toJS();
  $: entryExists =
    $checkForSourceInCatalog?.error?.response?.data?.message !==
    "entry not found";

  $: connectors = useRuntimeServiceListConnectors();

  // get the connector for this source type, if valid
  $: currentConnector = $connectors?.data?.connectors?.find(
    (connector) => connector?.name === source?.type
  );
  $: allConnectors = $connectors?.data?.connectors?.map(
    (connector) => connector.name
  );
  $: remoteConnectorNames = allConnectors
    ?.map((connector) => connector.name)
    ?.filter((name) => name !== "local_file");

  const refreshSourceMutation = useRuntimeServiceRefreshAndReconcile();
  const createSource = useRuntimeServicePutFileAndReconcile();
  const queryClient = useQueryClient();

  let uploadErrors = undefined;
  $: uploadErrors = $fileArtifactsStore.entities[sourcePath]?.errors;
  const onRefreshClick = async (tableName: string) => {
    try {
      await refreshSource(
        currentConnector?.name,
        tableName,
        $runtimeStore?.instanceId,
        $refreshSourceMutation,
        $createSource,
        queryClient
      );
      const queryKey = getRuntimeServiceGetCatalogEntryQueryKey(
        $runtimeStore?.instanceId,
        tableName
      );
      await queryClient.refetchQueries(queryKey);
    } catch (err) {
      // no-op
    }
    overlay.set(null);
  };

  $: isReconciling = $fileArtifactsStore.entities[sourcePath]?.isReconciling;
</script>

{#key sourceName}
  <WorkspaceContainer assetID={sourceName}>
    <div slot="header">
      <SourceWorkspaceHeader {sourceName} />
    </div>
    <div
      slot="body"
      class="grid pb-3"
      style:grid-template-rows="max-content auto"
      style:height="100vh"
    >
      {#if entryExists}
        <div
          style:overflow="auto"
          style:height="calc(100vh - var(--header-height) - 2rem)"
          class="m-4 border border-gray-300 rounded"
        >
          {#key sourceName}
            <ConnectedPreviewTable objectName={sourceName} />
          {/key}
        </div>
      {:else}
        <!-- error states -->
        <div
          class="errors flex flex-col items-center pt-8 gap-y-4 m-auto mt-0 text-gray-500"
          style:width="500px"
        >
          {#if !allConnectors.includes(source?.type)}
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
              <a href="https://docs.rilldata.com/using-rill/import-data"
                >view the documentation</a
              >.
            </div>
          {:else if source?.type === "local_file"}
            <div class="text-center">
              The data file for <span class="font-bold">{sourceName}</span> has not
              been imported as a source.
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
              is empty. Edit <b>{`sources/${sourceName}.yaml`}</b> to add a source
              definition.
            </div>
            <div>
              For more information,
              <a href="https://docs.rilldata.com/using-rill/import-data"
                >view the documentation</a
              >.
            </div>
          {:else if !source?.type}
            <div>
              The source <span class="font-bold">{sourceName}</span> does not
              have a defined type. Edit <b>{`sources/${sourceName}.yaml`}</b> to
              add "type:
              {"<filetype>"}"
            </div>
            <div>
              For more information,
              <a href="https://docs.rilldata.com/using-rill/import-data"
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
              <a href="https://docs.rilldata.com/using-rill/import-data"
                >view the documentation</a
              >.
            </div>
          {:else if isReconciling}
            <div class="text-center">
              The source <span class="font-bold">{sourceName}</span> is being imported.
            </div>
            <div class="text-center">This may take a few seconds.</div>
          {:else}
            <div class="text-center">
              The source <span class="font-bold">{sourceName}</span> has not been
              imported.
            </div>
            <Button
              type="primary"
              on:click={async () => {
                uploadErrors = undefined;
                await onRefreshClick(sourceName);
              }}
              >Import data
            </Button>
          {/if}
          <!-- show any remaining errors -->
          {#if uploadErrors?.length}
            <Callout level="error">
              {#each uploadErrors as error}
                {humanReadableErrorMessage(
                  currentConnector?.name,
                  3,
                  error.message
                )}
              {/each}
            </Callout>
          {/if}
        </div>
      {/if}
    </div>

    <SourceInspector {sourceName} slot="inspector" />
  </WorkspaceContainer>
{/key}

<style>
  .errors > div:not(.text-center) {
    text-align: left;
    width: 500px;
  }
</style>
