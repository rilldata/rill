<!-- Non-destructive confirmation dialog for superuser actions -->
<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let open = false;
  export let title: string;
  export let description: string;
  export let confirmLabel: string = "Confirm";
  export let loading = false;
  export let onConfirm: () => Promise<void>;

  let confirming = false;

  async function handleConfirm() {
    confirming = true;
    try {
      await onConfirm();
      open = false;
    } catch {
      // Keep dialog open for retry
    } finally {
      confirming = false;
    }
  }

  $: isLoading = loading || confirming;
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>
      <AlertDialogDescription>{description}</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        large
        class="font-normal"
        type="tertiary"
        onClick={() => (open = false)}
      >
        Cancel
      </Button>
      <Button
        large
        class="font-normal"
        type="primary"
        onClick={handleConfirm}
        loading={isLoading}
      >
        {confirmLabel}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
