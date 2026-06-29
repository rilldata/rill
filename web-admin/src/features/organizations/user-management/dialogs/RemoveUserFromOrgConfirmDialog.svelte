<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
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
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent noCancel>
    <AlertDialogHeader>
      <AlertDialogTitle>{m.users_remove_confirm_title()}</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {m.users_remove_confirm_desc()}
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
      <Button type="destructive" onClick={handleRemove}>{m.users_yes_remove()}</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
