<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog/index";
  import { Button } from "@rilldata/web-common/components/button";

  export let open: boolean;
  export let onDelete: () => void;

  function handleClose() {
    open = false;
  }
</script>

<AlertDialog.Root
  {open}
  onOpenChange={(open) => {
    if (!open) {
      handleClose();
    }
  }}
>
  <AlertDialog.Content>
    <AlertDialog.Title>
      Delete all files in this folder?
    </AlertDialog.Title>

    <AlertDialog.Description>
      This folder is not empty. All contained items will be deleted.
    </AlertDialog.Description>

    <AlertDialog.Footer>
      <AlertDialog.Action>
        {#snippet child({ props })}
          <Button
            {...props}
            large
            onClick={() => {
              handleClose();
              onDelete();
            }}
            type="destructive"
          >
            Delete
          </Button>
        {/snippet}
      </AlertDialog.Action>

      <AlertDialog.Cancel>
        {#snippet child({ props })}
          <Button {...props} large onClick={handleClose} type="tertiary"
            >Cancel</Button
          >
        {/snippet}
      </AlertDialog.Cancel>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
