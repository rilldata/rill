<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import AddDataForm from "@rilldata/web-common/features/sources/modal/AddDataForm.svelte";
  import LocalSourceUpload from "@rilldata/web-common/features/sources/modal/LocalSourceUpload.svelte";
  import { createRuntimeServiceListConnectorDrivers } from "@rilldata/web-common/runtime-client";
  import type { PageData } from "./$types";

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

  async function onContinue(filePath: string) {
    if ($managementType === "rill-managed") {
      onboardingState.complete();
      // Navigate to the new source (whether there's an error or not)
      await goto(`/files/${filePath}`);
    } else {
      // Continue in the onboarding wizard to create a dashboard
      await goto(`/welcome/make-your-first-dashboard`);
    }
  }

  function onBack() {
    try {
      onboardingState.cleanUp();
    } catch (e) {
      console.error(e);
    }
  }
</script>

{#if $firstDataSource === "local_file"}
  <LocalSourceUpload
    on:close={onContinue}
    on:back={onBack}
    backHref="/welcome/select-connectors"
  />
{:else if connectorDriver}
  <div class="w-[544px] p-6 overflow-visible">
    <h2 class="text-lead pb-2">Connect to {connectorDriver.displayName}</h2>
    <AddDataForm
      formType={$managementType === "self-managed" ? "connector" : "source"}
      connector={connectorDriver}
      olapDriver={$olapDriver}
      onSuccess={onContinue}
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
