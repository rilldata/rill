<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { Dialog } from "./index";

  export let title: string;
  export let description: string;
  export let confirmLabel: string;
  export let cancelLabel: string;
  export let open = false;

  let showCancelDialog = false;
  function onCancel() {
    showCancelDialog = true;
  }

  function onConfirmCancel() {
    open = false;
    showCancelDialog = false;
  }
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    onCancel();
  }}
  onOpenChange={(o) => {
    // Hack to intercept cancel from clicking X or pressing escape
    if (!o) onCancel();
    setTimeout(() => (open = true));
  }}
>
  <slot {onCancel} />
</Dialog>

<AlertDialog bind:open={showCancelDialog}>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>
      <AlertDialogDescription>
        {description}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="secondary" on:click={() => (showCancelDialog = false)}>
        {cancelLabel}
      </Button>
      <Button type="primary" on:click={onConfirmCancel}>
        {confirmLabel}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
