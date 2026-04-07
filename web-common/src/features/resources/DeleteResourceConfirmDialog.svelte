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

  let {
    open = $bindable(false),
    resourceKind,
    title,
    description,
    onDelete,
  }: {
    open: boolean;
    resourceKind: string;
    title: string;
    description: string;
    onDelete: () => Promise<void>;
  } = $props();

  async function handleDelete() {
    try {
      await onDelete();
      open = false;
    } catch (error) {
      console.error(`Failed to delete ${resourceKind}:`, error);
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
      <AlertDialogTitle>Delete this {resourceKind}?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          The {resourceKind} "<strong>{title}</strong>" will be permanently
          deleted {description}.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="destructive" onClick={handleDelete}>Yes, delete</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
