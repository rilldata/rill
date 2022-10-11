<script lang="ts">
  import Tab from "@rilldata/web-local/lib/components/tab/Tab.svelte";
  import TabGroup from "@rilldata/web-local/lib/components/tab/TabGroup.svelte";
  import { createEventDispatcher } from "svelte";
  import { Dialog } from "../../modal";
  import LocalSource from "./LocalSource.svelte";
  import RemoteSource from "./RemoteSource.svelte";

  const dispatch = createEventDispatcher();

  let selectedTab = "remote";
  let disabled = false;
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
    <div>
      <TabGroup
        on:select={(event) => {
          selectedTab = event.detail;
        }}
      >
        <Tab value={"remote"}>Remote source</Tab>
        <Tab value={"local"}>Local source</Tab>
        <!-- <Tab value={"example"}>Example source</Tab> -->
      </TabGroup>
    </div>
  </div>
  <div class="overflow-y-auto flex-grow">
    {#if selectedTab === "remote"}
      <RemoteSource on:cancel={() => dispatch("close")} />
    {:else if selectedTab === "local"}
      <LocalSource on:close />
      <!-- {:else if selectedTab === "example"}
      <ExampleSource /> -->
    {/if}
  </div>
</Dialog>
