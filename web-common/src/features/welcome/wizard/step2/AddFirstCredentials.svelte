<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import { createRuntimeServiceListConnectorDrivers } from "../../../../runtime-client";
  import type { OlapDriver } from "../../../connectors/olap/olap-config";
  import AddDataForm from "../../../sources/modal/AddDataForm.svelte";
  import LocalSourceUpload from "../../../sources/modal/LocalSourceUpload.svelte";

  export let managementType: "self-managed" | "rill-managed";
  export let olapDriver: OlapDriver;
  export let firstDataSource: string | undefined;
  export let onBack: () => void;
  export let onContinue: (filePath: string) => void;

  // Get connector driver
  const connectorsQuery = createRuntimeServiceListConnectorDrivers();
  $: firstConnectorName =
    managementType === "rill-managed" ? firstDataSource : olapDriver;
  $: connectorDriver = $connectorsQuery.data?.connectors?.find(
    (c) => c.name === firstConnectorName,
  );
</script>

{#if firstDataSource === "local_file"}
  <LocalSourceUpload on:close={onContinue} on:back={onBack} />
{:else if connectorDriver}
  <div class="w-[544px] p-6">
    <h2 class="text-lead">Connect to {connectorDriver.displayName}</h2>
    <AddDataForm
      formType={managementType === "self-managed" ? "connector" : "source"}
      connector={connectorDriver}
      {olapDriver}
      onSubmit={onContinue}
    >
      <svelte:fragment slot="actions" let:submitting>
        <div class="flex flex-col gap-y-2">
          <Button
            type="primary"
            form="add-data-form"
            submitForm
            disabled={submitting}
            large
          >
            {submitting
              ? "Testing connection..."
              : firstDataSource
                ? "Add data"
                : "Connect"}
          </Button>
          <Button type="link" large on:click={onBack}>Back</Button>
        </div>
      </svelte:fragment>
    </AddDataForm>
  </div>
{/if}
