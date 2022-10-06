<script lang="ts">
  import Tab from "@rilldata/web-local/lib/components/tab/Tab.svelte";
  import TabGroup from "@rilldata/web-local/lib/components/tab/TabGroup.svelte";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../button";
  import { Dialog } from "../../modal";
  import ExampleSource from "./ExampleSource.svelte";
  import LocalSource from "./LocalSource.svelte";
  import RemoteSource from "./RemoteSource.svelte";

  const dispatch = createEventDispatcher();

  let selectedTab = "remote";
  let remoteSourceSelectedConnector;
  let disabled = false;
</script>

<Dialog
  yFixed
  size="lg"
  showCancel
  {disabled}
  on:cancel={() => dispatch("close")}
>
  <div slot="title">
    <TabGroup
      on:select={(event) => {
        selectedTab = event.detail;
      }}
    >
      <Tab value={"remote"}>Remote source</Tab>
      <Tab value={"local"}>Local source</Tab>
      <Tab value={"example"}>Example source</Tab>
    </TabGroup>
  </div>
  <svelte:fragment slot="body">
    {#if selectedTab === "remote"}
      <RemoteSource
        on:select-connector={(event) => {
          remoteSourceSelectedConnector = event.detail;
        }}
      />
    {:else if selectedTab === "local"}
      <LocalSource />
    {:else if selectedTab === "example"}
      <ExampleSource />
    {/if}
  </svelte:fragment>
  <svelte:fragment slot="footer">
    {#if selectedTab === "remote"}
      <Button
        type="primary"
        submitForm
        form="remote-source-{remoteSourceSelectedConnector}-form"
      >
        Add source
      </Button>
    {/if}
  </svelte:fragment>
</Dialog>
