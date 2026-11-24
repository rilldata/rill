<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import CreateTokenForm from "./CreateTokenForm.svelte";

  let { open = $bindable(false) }: { open?: boolean } = $props();
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-w-lg">
    <CreateTokenForm onDone={() => (open = false)} />
  </Dialog.Content>
</Dialog.Root>
