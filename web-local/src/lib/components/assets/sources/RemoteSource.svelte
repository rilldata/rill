<script lang="ts">
  import Tab from "../../tab/Tab.svelte";
  import TabGroup from "../../tab/TabGroup.svelte";
  import RemoteSourceForm from "./RemoteSourceForm.svelte";

  let selectedConnector = "S3";
  let selectedConnectorDescription = "";
</script>

<div class="h-full flex flex-col">
  <div class="pt-4 px-4">
    <TabGroup
      variant="secondary"
      on:select={(event) => {
        selectedConnector = event.detail;
      }}
    >
      <Tab value={"S3"}>S3</Tab>
      <Tab value={"GCS"}>GCS</Tab>
      <Tab value={"HTTP"}>https</Tab>
    </TabGroup>
  </div>
  <div class="p-4">
    {@html selectedConnectorDescription}
  </div>
  {#key selectedConnector}
    <RemoteSourceForm
      connector={selectedConnector}
      bind:connectorDescription={selectedConnectorDescription}
    />
  {/key}
</div>
