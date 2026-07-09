<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{m.users_upgrade_confirm_title({ role: newRole })}</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {m.users_upgrade_confirm_desc({ role: newRole })}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>{m.users_cancel()}</Button
      >
      <Button type="primary" onClick={handleUpgrade}>{m.users_yes_upgrade()}</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
