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
  export let newRole: string;
  export let onUpgrade: (email: string, role: string) => void;

  async function handleUpgrade() {
    try {
      onUpgrade(email, newRole);
      open = false;
    } catch (error) {
      console.error("Failed to upgrade user role:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Upgrade guest to {newRole}?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          Upgrading a guest to {newRole} will grant this user access to all open
          projects in the organization. Would you like to upgrade this guest user
          to {newRole}?
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
      <Button type="primary" onClick={handleUpgrade}>Yes, upgrade</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
