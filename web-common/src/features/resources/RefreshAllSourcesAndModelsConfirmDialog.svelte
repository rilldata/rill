<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let open = false;
  export let onRefresh: () => void;

  function handleRefresh() {
    try {
      onRefresh();
      open = false;
    } catch (error) {
      console.error("Failed to refresh resource:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{m.status_refresh_all_confirm_title()}</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {m.status_refresh_all_confirm_body()}
          <br />
          <br />
          <span class="font-medium">{m.status_note()}</span> {m.status_refresh_all_confirm_tip()}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>{m.status_cancel()}</Button
      >
      <Button type="primary" onClick={handleRefresh}>{m.status_yes_refresh()}</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
