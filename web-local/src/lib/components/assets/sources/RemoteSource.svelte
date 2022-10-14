<script lang="ts">
  import {
    useRuntimeServiceListConnectors,
    V1Connector,
  } from "web-common/src/runtime-client";
  import Tab from "../../tab/Tab.svelte";
  import TabGroup from "../../tab/TabGroup.svelte";
  import RemoteSourceForm from "./RemoteSourceForm.svelte";

  let selectedConnector: V1Connector;

  const connectors = useRuntimeServiceListConnectors({
    // remove local "file" connector
    query: {
      select: (data) => {
        data.connectors =
          data.connectors &&
          data.connectors.filter((connector) => connector.name !== "file");
        return data;
      },
    },
  });

  function setDefaultConnector(connectors: V1Connector[]) {
    if (connectors?.length > 0) {
      selectedConnector = connectors[0];
    }
  }

  $: setDefaultConnector($connectors.data.connectors);
</script>

<div class="h-full flex flex-col">
  <div class="pt-4 px-4">
    <TabGroup
      variant="secondary"
      on:select={(event) => {
        selectedConnector = event.detail;
      }}
    >
      {#if $connectors.isSuccess && $connectors.data && $connectors.data.connectors?.length > 0}
        {#each $connectors.data.connectors as connector}
          <Tab value={connector}>{connector.displayName}</Tab>
        {/each}
      {/if}
    </TabGroup>
  </div>
  <div class="p-4">
    {@html selectedConnector?.description}
  </div>
  {#if selectedConnector}
    {#key selectedConnector}
      <RemoteSourceForm connector={selectedConnector} on:close />
    {/key}
  {/if}
</div>
