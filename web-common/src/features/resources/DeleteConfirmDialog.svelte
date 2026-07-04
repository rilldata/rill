<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";

  export let open = false;
  export let title: string;
  export let onDelete: () => Promise<void>;

  async function handleDelete() {
    try {
      await onDelete();
      open = false;
    } catch (error) {
      console.error("Delete failed:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{title}</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          <slot />
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}
      >
        Cancel
      </Button>
      <Button type="destructive" onClick={handleDelete}>Yes, delete</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
