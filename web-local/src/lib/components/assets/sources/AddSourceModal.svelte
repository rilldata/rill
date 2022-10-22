<script lang="ts">
  import Tab from "@rilldata/web-local/lib/components/tab/Tab.svelte";
  import TabGroup from "@rilldata/web-local/lib/components/tab/TabGroup.svelte";
  import { createEventDispatcher } from "svelte";
  import {
    useRuntimeServiceListConnectors,
    V1Connector,
  } from "web-common/src/runtime-client";
  import { Dialog } from "../../modal";
  import LocalSource from "./LocalSource.svelte";
  import RemoteSource from "./RemoteSource.svelte";

  const dispatch = createEventDispatcher();

  let selectedConnector: V1Connector;

  const TAB_ORDER = ["gcs", "s3", "file"];

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
  yFixed
  size="lg"
  showCancel
  compact
  {disabled}
  on:cancel={() => dispatch("close")}
>
  <div slot="title">
    <TabGroup
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
  {#if selectedConnector?.name !== "file"}
    <div class="p-4">
      {@html selectedConnector?.description}
    </div>
  {/if}
  <div class="h-full flex flex-col">
    {#if selectedConnector?.name === "gcs" || selectedConnector?.name === "s3"}
      {#key selectedConnector}
        <RemoteSource connector={selectedConnector} on:close />
      {/key}
    {/if}
    {#if selectedConnector?.name === "file"}
      <LocalSource on:close />
    {/if}
  </div>
</Dialog>
