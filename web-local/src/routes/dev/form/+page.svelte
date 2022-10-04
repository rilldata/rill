<script lang="ts">
  import Tab from "@rilldata/web-local/lib/components/tab/Tab.svelte";
  import TabGroup from "@rilldata/web-local/lib/components/tab/TabGroup.svelte";
  import ConnectorForm from "../../../lib/components/form/ConnectorForm.svelte";
  import {
    GCS,
    GCSYupSchema,
    HTTP,
    HTTPYupSchema,
    S3,
    S3YupSchema,
  } from "../../../lib/connectors/schemas";

  let sourceName = "my_new_source";

  let connectorOptions = [
    { label: "S3", value: "S3" },
    { label: "GCS", value: "GCS" },
    { label: "HTTP", value: "HTTP" },
  ];

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
  on:select={(event) => {
    selectedConnector = event.detail;
  }}
>
  <Tab value={"S3"}>S3</Tab>
  <Tab value={"GCS"}>GCS</Tab>
  <Tab value={"HTTP"}>https</Tab>
</TabGroup>

<div class="py-8">
  {#key selectedConnector}
    <ConnectorForm {sourceName} {connectorSpec} {yupSchema} />
  {/key}
</div>
