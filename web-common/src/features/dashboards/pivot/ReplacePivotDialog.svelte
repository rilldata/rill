<script>
  import { Button } from "@rilldata/web-common/components/button";
  import { createEventDispatcher } from "svelte";
  import Dialog from "../../../components/dialog-v2/Dialog.svelte";

  export let open;

  const dispatch = createEventDispatcher();
  function onClose() {
    dispatch("close");
  }
  function onReplace() {
    dispatch("replace");
  }
</script>

<Dialog titleMarginBottomOverride="mb-4" on:close {open}>
  <svelte:fragment slot="title">Replace current pivot table?</svelte:fragment>

  <div class="flex flex-col items-center" slot="body">
    <div class="text-left text-sm text-gray-500 w-full">
      Starting a new table will loose your previous work. Bookmark tables you
      want to keep
    </div>
  </div>

  <svelte:fragment slot="footer">
    <div class="flex gap-x-2">
      <div class="grow" />
      <Button type="secondary" on:click={onClose}>Close</Button>
      <Button type="brand" on:click={onReplace}>Replace</Button>
    </div>
  </svelte:fragment>
</Dialog>
