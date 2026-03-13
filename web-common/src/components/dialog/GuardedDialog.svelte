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
  import AlertCircle from "@rilldata/web-common/components/icons/AlertCircle.svelte";
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

  function onClose() {
    open = false;
  }

  function onConfirmCancel() {
    open = false;
    showCancelDialog = false;
  }
</script>

<!-- Dialog with confirm on cancel need a strong intent. We intercept all close
     attempts via onOpenChange and show a cancel confirmation dialog instead. -->
<Dialog
  bind:open
  onOpenChange={(o) => {
    // Intercept cancel from clicking X, pressing escape, or clicking outside
    if (!o && open) onCancel();
    setTimeout(() => (open = true));
  }}
>
  <slot {onCancel} {onClose} />
</Dialog>

<AlertDialog bind:open={showCancelDialog}>
  <AlertDialogTrigger>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <div class="flex flex-row items-center gap-x-2">
        <AlertCircle size="20px" className="text-yellow-500" />
        <AlertDialogTitle>{title}</AlertDialogTitle>
      </div>
      <AlertDialogDescription class="pl-7">
        {description}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="secondary" onClick={() => (showCancelDialog = false)}>
        {cancelLabel}
      </Button>
      <Button type="primary" onClick={onConfirmCancel}>
        {confirmLabel}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
