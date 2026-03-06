<script lang="ts">
  import { pushState } from "$app/navigation";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import AddData from "@rilldata/web-common/features/add-data/AddData.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
  import { transitionToNextStep } from "@rilldata/web-common/features/add-data/steps/transitions.ts";
  import { AddDataStep } from "@rilldata/web-common/features/add-data/steps/types.ts";

  export let open: boolean;
  export let schema: string | undefined = undefined;
  export let connector: string | undefined = undefined;

  $: ({ instanceId } = $runtime);

  $: config = { instanceId, importOnly: true };
  $: initArgs = { schema, connector };

  $: if (open) {
    pushState(
      "",
      transitionToNextStep(config, { step: AddDataStep.Select }, initArgs),
    );
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="p-0 w-[900px] max-w-[900px] h-[600px]">
    <AddData config={{ instanceId, importOnly: true }} />
  </Dialog.Content>
</Dialog.Root>
