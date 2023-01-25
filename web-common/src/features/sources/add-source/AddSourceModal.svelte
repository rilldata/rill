<script lang="ts">
  import { Dialog } from "@rilldata/web-common/components/modal";
  import Tab from "@rilldata/web-common/components/tab/Tab.svelte";
  import TabGroup from "@rilldata/web-common/components/tab/TabGroup.svelte";
  import {
    useRuntimeServiceListConnectors,
    V1Connector,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import LocalSource from "./LocalSource.svelte";
  import RemoteSource from "./RemoteSource.svelte";

  const dispatch = createEventDispatcher();

  let showLocalFileDetailedOptions;
  let selectedConnector: V1Connector;

  const TAB_ORDER = ["gcs", "s3", "https", "local_file"];

  const connectors = useRuntimeServiceListConnectors({
    query: {
      // arrange connectors in the way we would like to display them
      select: (data) => {
        data.connectors =
          data.connectors &&
          data.connectors.sort(
            (a, b) => TAB_ORDER.indexOf(a.name) - TAB_ORDER.indexOf(b.name)
          );
        return data;
      },
    },
  });

  let disabled = false;

  function setDefaultConnector(connectors: V1Connector[]) {
    if (connectors?.length > 0) {
      selectedConnector = connectors[0];
    }
  }

  $: setDefaultConnector($connectors.data.connectors);
</script>

<Dialog
  compact
  useContentForMinSize
  {disabled}
  on:cancel={() => dispatch("close")}
  showCancel
  size="md"
  yFixed
>
  <div slot="title">
    <TabGroup
      on:select={(event) => {
        selectedConnector = event.detail;
      }}
    >
      {#if $connectors.isSuccess && $connectors.data && $connectors.data.connectors?.length > 0}
        {#each $connectors.data.connectors as connector}
          <Tab selected={selectedConnector === connector} value={connector}
            >{connector.displayName}</Tab
          >
        {/each}
      {/if}
    </TabGroup>
  </div>
  <div class="flex-grow overflow-y-auto">
    {#if selectedConnector?.name === "local_file" && !showLocalFileDetailedOptions}
      <LocalSource
        bind:showDetailedOptions={showLocalFileDetailedOptions}
        on:close
      />
    {/if}
    {#if selectedConnector?.name === "gcs" || selectedConnector?.name === "s3" || selectedConnector?.name === "https" || showLocalFileDetailedOptions}
      {#key selectedConnector}
        <RemoteSource connector={selectedConnector} on:close />
      {/key}
    {/if}
  </div>
</Dialog>
