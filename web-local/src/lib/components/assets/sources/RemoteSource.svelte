<script lang="ts">
  import {
    GCS,
    GCSYupSchema,
    HTTP,
    HTTPYupSchema,
    S3,
    S3YupSchema,
  } from "../../../connectors/schemas";
  import Tab from "../../tab/Tab.svelte";
  import TabGroup from "../../tab/TabGroup.svelte";
  import RemoteSourceForm from "./RemoteSourceForm.svelte";

  let sourceName = "my_new_source";

  let selectedConnector = "S3";
  let connectorSpec = S3;
  let yupSchema = S3YupSchema;

  function onChange(selectedConnector: string) {
    if (selectedConnector === "S3") {
      connectorSpec = S3;
      yupSchema = S3YupSchema;
    } else if (selectedConnector === "GCS") {
      connectorSpec = GCS;
      yupSchema = GCSYupSchema;
    } else if (selectedConnector === "HTTP") {
      connectorSpec = HTTP;
      yupSchema = HTTPYupSchema;
    }
  }

  $: onChange(selectedConnector);
</script>

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

<div class="pt-8">
  {#key selectedConnector}
    <RemoteSourceForm {sourceName} {connectorSpec} {yupSchema} />
  {/key}
</div>
