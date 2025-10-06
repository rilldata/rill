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
  import Button from "web-common/src/components/button/Button.svelte";

  export let open = false;
  export let email: string;
  export let onRemove: (email: string) => void;

  async function handleRemove() {
    try {
      onRemove(email);
      open = false;
    } catch (error) {
      console.error("Failed to remove user from organization:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent noCancel>
    <AlertDialogHeader>
      <AlertDialogTitle>Remove user from organization?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This user will no longer be able to access the organization.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" status="error" onClick={handleRemove}
        >Yes, remove</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
