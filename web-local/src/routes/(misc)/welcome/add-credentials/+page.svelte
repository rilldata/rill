<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import AddDataForm from "@rilldata/web-common/features/sources/modal/AddDataForm.svelte";
  import LocalSourceUpload from "@rilldata/web-common/features/sources/modal/LocalSourceUpload.svelte";
  import { createRuntimeServiceListConnectorDrivers } from "@rilldata/web-common/runtime-client";

  export let data: PageData;
  const { onboardingState } = data;
  const { managementType, olapDriver, firstDataSource } = onboardingState;

  // Get connector driver
  const connectorsQuery = createRuntimeServiceListConnectorDrivers();
  $: firstConnectorName =
    $managementType === "rill-managed" ? $firstDataSource : $olapDriver;
  $: connectorDriver = $connectorsQuery.data?.connectors?.find(
    (c) => c.name === firstConnectorName,
  );

  function onContinue(filePath: string) {
    console.log("onContinue", filePath);
  }

  function onBack() {
    onboardingState.cleanUp();
  }
</script>

{#if $firstDataSource === "local_file"}
  <LocalSourceUpload on:close={onContinue} on:back={onBack} />
{:else if connectorDriver}
  <div class="w-[544px] p-6">
    <h2 class="text-lead">Connect to {connectorDriver.displayName}</h2>
    <AddDataForm
      formType={$managementType === "self-managed" ? "connector" : "source"}
      connector={connectorDriver}
      olapDriver={$olapDriver}
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
              : $firstDataSource
                ? "Add data"
                : "Connect"}
          </Button>
          <Button
            type="link"
            large
            href="/welcome/select-connectors"
            on:click={onBack}
          >
            Back
          </Button>
        </div>
      </svelte:fragment>
    </AddDataForm>
  </div>
{/if}
