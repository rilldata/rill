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
    branch,
    onConfirm,
  }: {
    open: boolean;
    branch: string;
    onConfirm: () => void;
  } = $props();
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete this branch?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          The branch <span class="font-mono text-xs font-medium">{branch}</span>
          will be deleted. Any unpushed changes will be lost.
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
      <Button
        type="destructive"
        onClick={() => {
          open = false;
          onConfirm();
        }}>Yes, delete</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
