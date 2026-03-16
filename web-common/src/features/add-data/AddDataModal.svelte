<script lang="ts">
  import { pushState } from "$app/navigation";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import AddData from "@rilldata/web-common/features/add-data/AddData.svelte";
  import { transitionToNextStep } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  export let open: boolean;
  export let schema: string | undefined = undefined;
  export let connector: string | undefined = undefined;

  const runtimeClient = useRuntimeClient();

  $: config = { runtimeClient, importOnly: true };
  $: initArgs = { schema, connector };

  async function transitionToInit() {
    pushState(
      "",
      await transitionToNextStep(
        runtimeClient,
        { step: AddDataStep.SelectConnector },
        initArgs,
      ),
    );
  }
  $: if (open) void transitionToInit();
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="p-0 w-fit max-w-fit h-fit">
    <AddData {config} />
  </Dialog.Content>
</Dialog.Root>
