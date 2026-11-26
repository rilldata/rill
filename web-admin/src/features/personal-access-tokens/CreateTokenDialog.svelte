<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import CreateTokenForm from "./CreateTokenForm.svelte";

  let { open = $bindable(false) }: { open?: boolean } = $props();

  // Closing mid-request would unmount the form while the server may still issue the token.
  let submitting = $state(false);
</script>

<!-- Outside clicks never close it: one could lose a token that hasn't been copied yet. -->
<Dialog.Root bind:open>
  <Dialog.Content
    class="max-w-lg"
    noClose={submitting}
    escapeKeydownBehavior={submitting ? "ignore" : "close"}
    interactOutsideBehavior="ignore"
  >
    <CreateTokenForm bind:submitting onDone={() => (open = false)} />
  </Dialog.Content>
</Dialog.Root>
