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
      Are you sure you want to delete everything?
    </AlertDialog.Title>

    <AlertDialog.Description>
      This folder is not empty. All contained items will be deleted.
    </AlertDialog.Description>

    <AlertDialog.Footer>
      <AlertDialog.Action asChild let:builder>
        <Button
          large
          builders={[builder]}
          onClick={() => {
            handleClose();
            onDelete();
          }}
          type="destructive"
        >
          Delete
        </Button>
      </AlertDialog.Action>

      <AlertDialog.Cancel asChild let:builder>
        <Button
          large
          builders={[builder]}
          onClick={handleClose}
          type="tertiary"
        >
          Cancel
        </Button>
      </AlertDialog.Cancel>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>
