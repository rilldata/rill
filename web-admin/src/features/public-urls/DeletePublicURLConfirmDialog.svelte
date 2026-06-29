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
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let open = false;
  export let id: string;
  export let onDelete: (id: string) => void;

  async function handleDelete() {
    try {
      onDelete(id);
      open = false;
    } catch (error) {
      console.error("Failed to delete public URL:", error);
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
      <AlertDialogTitle>{m.public_url_delete_title()}</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {m.public_url_delete_description()}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>{m.public_url_cancel_button()}</Button
      >
      <Button type="destructive" onClick={handleDelete}>{m.public_url_yes_delete_button()}</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
