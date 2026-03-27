<script lang="ts">
  import AddDataManager from "@rilldata/web-common/features/add-data/manager/AddDataManager.svelte";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { WelcomeStatus } from "@rilldata/web-common/features/welcome/status.ts";

  let addDataStep: AddDataStep = AddDataStep.SelectConnector;

  $: isImportStep = addDataStep === AddDataStep.Import;
</script>

<div class:my-auto={isImportStep}>
  {#if !isImportStep}
    <div class="text-base font-semibold text-fg-secondary">Getting started</div>
    <div class="text-3xl font-bold text-fg-accent">Connect your data</div>
  {/if}
  <div class="w-fit h-fit mt-4">
    <AddDataManager
      config={{ welcomeScreen: true }}
      onStepChange={(step) => (addDataStep = step)}
      onClose={() => WelcomeStatus.set(false)}
    />
  </div>
</div>
