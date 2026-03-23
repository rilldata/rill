<script lang="ts">
  import { pushState, replaceState } from "$app/navigation";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import AddData from "@rilldata/web-common/features/add-data/AddData.svelte";
  import { transitionToNextStep } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import {
    type AddDataConfig,
    AddDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let open: boolean = true;
  export let schema: string | undefined = undefined;
  export let connector: string | undefined = undefined;

  const runtimeClient = useRuntimeClient();

  const config: AddDataConfig = { importOnly: true };

  $: if (open) void transitionToInit();

  async function transitionToInit() {
    pushState(
      "",
      await transitionToNextStep(
        runtimeClient,
        { step: AddDataStep.SelectConnector },
        { schema, connector },
      ),
    );
  }

  function maybeReplaceState(newOpen: boolean) {
    if (!newOpen) {
      replaceState("", {});
    }
  }
</script>

<Dialog.Root bind:open onOpenChange={maybeReplaceState}>
  <Dialog.Content class="p-0 w-fit max-w-fit h-fit" noClose>
    {#if open}
      <AddData {config} onClose={() => (open = false)} />
    {/if}
  </Dialog.Content>
</Dialog.Root>
