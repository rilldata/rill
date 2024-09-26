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
  export let groupName: string;
  export let onDelete: (groupName: string) => void;

  async function handleDelete() {
    try {
      onDelete(groupName);
      open = false;
    } catch (error) {
      console.error("Failed to delete user group:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete this user group?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This user group will no longer be able to access the organization.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" status="error" on:click={handleDelete}
        >Yes, delete</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
