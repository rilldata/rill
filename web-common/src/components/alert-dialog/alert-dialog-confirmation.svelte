<script lang="ts">
  import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import {
    Button,
    type ButtonType,
  } from "@rilldata/web-common/components/button/index.js";

  // A simple confirm/cancel alert dialog. Open state is controlled by the caller
  // (open + onOpenChange) so it can be driven by any kind of pending-action state.
  export let open = false;
  export let onOpenChange: (open: boolean) => void = () => {};
  export let title: string;
  export let description: string;
  export let confirmLabel = "Confirm";
  export let confirmType: ButtonType = "primary";
  export let onConfirm: () => void;
</script>

<AlertDialog {open} {onOpenChange}>
  <AlertDialogContent>
    <AlertDialogTitle>{title}</AlertDialogTitle>
    <AlertDialogDescription>{description}</AlertDialogDescription>
    <AlertDialogFooter>
      <AlertDialogCancel>
        {#snippet child({ props })}
          <Button {...props} large type="secondary">Cancel</Button>
        {/snippet}
      </AlertDialogCancel>
      <AlertDialogAction>
        {#snippet child({ props })}
          <Button {...props} large type={confirmType} onClick={onConfirm}>
            {confirmLabel}
          </Button>
        {/snippet}
      </AlertDialogAction>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
