<script lang="ts">
  import {
    useRuntimeServiceListConnectors,
    V1Connector,
  } from "@rilldata/web-common/runtime-client";
  import Tab from "@rilldata/web-local/lib/components/tab/Tab.svelte";
  import TabGroup from "@rilldata/web-local/lib/components/tab/TabGroup.svelte";
  import { createEventDispatcher } from "svelte";
  import { Dialog } from "../../modal";
  import LocalSource from "./LocalSource.svelte";
  import RemoteSource from "./RemoteSource.svelte";

  const dispatch = createEventDispatcher();

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
    {#if selectedConnector?.name === "gcs" || selectedConnector?.name === "s3" || selectedConnector?.name === "https"}
      {#key selectedConnector}
        <RemoteSource connector={selectedConnector} on:close />
      {/key}
    {/if}
    {#if selectedConnector?.name === "local_file"}
      <LocalSource on:close />
    {/if}
  </div>
</Dialog>
