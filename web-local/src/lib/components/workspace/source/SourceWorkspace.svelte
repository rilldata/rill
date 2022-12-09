<script lang="ts">
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
    useRuntimeServiceListConnectors,
  } from "@rilldata/web-common/runtime-client";
  import { appStore } from "@rilldata/web-local/lib/application-state-stores/app-store";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { parseDocument } from "yaml";

  import { EntityType } from "../../../../common/data-modeler-state-service/entity-state-service/EntityStateService";
  import Button from "../../button/Button.svelte";
  import { ConnectedPreviewTable } from "../../preview-table";
  import WorkspaceContainer from "../core/WorkspaceContainer.svelte";
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

  $: getSource = useRuntimeServiceGetFile(
    $runtimeStore?.instanceId,
    `sources/${sourceName}.yaml`
  );

  $: source = parseDocument($getSource?.data?.blob || "{}").toJS();
  $: entryExists =
    $checkForSourceInCatalog?.error?.response?.data?.message !==
    "entry not found";

  $: connectors = useRuntimeServiceListConnectors({
    query: {
      // arrange connectors in the way we would like to display them
      select: (data) => {
        data.connectors = data.connectors;
        return data;
      },
    },
  });

  // get the connector for this source type, if valid
  $: currentConnector = $connectors?.data?.connectors?.find(
    (connector) => connector.name === source?.type
  );
</script>

{#key sourceName}
  <WorkspaceContainer assetID={sourceName}>
    <div
      slot="body"
      class="grid pb-6"
      style:grid-template-rows="max-content auto"
      style:height="100vh"
    >
      <SourceWorkspaceHeader {sourceName} />
      {#if entryExists}
        <div
          style:overflow="auto"
          style:height="100%"
          class="m-6 mt-0 border border-gray-300 rounded"
        >
          {#key sourceName}
            <ConnectedPreviewTable objectName={sourceName} />
          {/key}
        </div>
      {:else}
        <!-- error states -->
        <div
          class="errors flex flex-col items-center pt-8 gap-y-4 m-auto mt-0"
          style:width="500px"
        >
          {#if source?.type === "file"}
            {#if !source?.path}
              <div class="text-gray-500">
                The source <span class="font-bold">{sourceName}</span> does not
                have a defined path. Edit <b>{`sources/${sourceName}.yaml`}</b>
                to add "path:
                {"<path>"}"
              </div>
              <div>
                For more information,
                <a href="https://docs.rilldata.com/using-rill/import-data"
                  >view the documentation</a
                >.
              </div>
            {:else}
              <div class="text-gray-500">
                The data file for <span class="font-bold">{sourceName}</span> has
                not been imported as a source.
              </div>
              <Button>Import file</Button>
            {/if}
          {:else if !Object.keys(source || {})?.length}
            <!-- source is empty -->
            <div class="text-gray-500">
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
            <div class="text-gray-500">
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
          {:else if ["gcs", "s3", "https"].includes(currentConnector.name) && !source?.uri}
            <div class="text-gray-500">
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
          {:else}
            <div class="text-gray-500 text-center">
              The source <span class="font-bold">{sourceName}</span> has not been
              imported.
            </div>
            <Button>Import data</Button>
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
