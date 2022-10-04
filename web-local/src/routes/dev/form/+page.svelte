<script lang="ts">
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

<div class="pb-6 flex flex-col gap-y-1">
  <h1 class="text-lg py-4">Connector</h1>
  {#each connectorOptions as option}
    <label for={option.value} class="flex gap-x-3">
      <input type="radio" bind:group={selectedConnector} value={option.value} />
      {option.label}
    </label>
  {/each}
</div>

{#key selectedConnector}
  <ConnectorForm {sourceName} {connectorSpec} {yupSchema} />
{/key}
