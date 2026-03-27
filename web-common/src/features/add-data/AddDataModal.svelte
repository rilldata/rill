<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import AddDataManager from "@rilldata/web-common/features/add-data/manager/AddDataManager.svelte";
  import { transitionToNextStep } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import {
    type AddDataConfig,
    type AddDataState,
    AddDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let open: boolean = true;
  export let schema: string | undefined = undefined;
  export let connector: string | undefined = undefined;

  const runtimeClient = useRuntimeClient();

  const config: AddDataConfig = { importOnly: true };
  // Use a boolean to mount remount when the modal is re-opened.
  // It is used to reset any state
  let showForm = false;
  let initStepState: AddDataState | undefined = undefined;

  $: if (open) void transitionToInit();

  async function transitionToInit() {
    initStepState = await transitionToNextStep(
      runtimeClient,
      { step: AddDataStep.SelectConnector },
      { schema, connector },
    );
    showForm = true;
  }
</script>

<Dialog.Root
  bind:open
  onOpenChange={(newOpen) => {
    if (!newOpen) showForm = false;
  }}
>
  <Dialog.Content class="p-0 w-fit max-w-fit h-fit" noClose>
    {#if showForm}
      <AddDataManager
        {config}
        {initStepState}
        onClose={() => {
          open = false;
          showForm = false;
        }}
      />
    {/if}
  </Dialog.Content>
</Dialog.Root>
