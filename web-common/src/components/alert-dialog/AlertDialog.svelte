<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  /**
   * Convenience component for a simple alert dialog
   */
  import {
    AlertDialog,
    AlertDialogTrigger,
    AlertDialogContent,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogDescription,
    AlertDialogFooter,
  } from "./index";

  export let title: string;
  export let description: string;
  export let confirmLabel: string;
  export let onConfirm: () => void;
  export let cancelLabel: string;
  export let open = false;

  function onCancel() {
    open = false;
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger><slot name="trigger" /></AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>
      <AlertDialogDescription>
        {description}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="secondary" on:click={onCancel}>
        {cancelLabel}
      </Button>
      <Button type="primary" on:click={onConfirm}>
        {confirmLabel}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
