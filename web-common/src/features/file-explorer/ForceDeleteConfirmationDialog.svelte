<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";

  export let open: boolean;
  export let onDelete: () => void;

  function handleClose() {
    open = false;
  }
</script>

<Dialog on:close={handleClose} {open}>
  <div slot="title">Are you sure you want to delete everything?</div>
  <div slot="body">
    This folder is not empty. All contained items will be deleted.
  </div>
  <div class="flex flex-row mt-4 gap-2" slot="footer">
    <div class="grow" />
    <Button on:click={handleClose} type="secondary">Cancel</Button>
    <Button
      on:click={() => {
        handleClose();
        onDelete();
      }}
      type="primary">Delete</Button
    >
  </div>
</Dialog>
