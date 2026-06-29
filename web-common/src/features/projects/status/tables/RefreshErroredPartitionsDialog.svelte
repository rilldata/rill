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
  export let modelName: string;
  export let onRefresh: () => Promise<void> | void;

  let isRefreshing = false;

  async function handleRefresh() {
    try {
      isRefreshing = true;
      await onRefresh();
      open = false;
    } catch (error) {
      console.error("Failed to refresh errored partitions:", error);
    } finally {
      isRefreshing = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>
        {m.status_refresh_errored_confirm_title({ modelName })}
      </AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {m.status_refresh_errored_confirm_body()}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        disabled={isRefreshing}
        onClick={() => {
          open = false;
        }}>{m.status_cancel()}</Button
      >
      <Button
        type="primary"
        onClick={handleRefresh}
        disabled={isRefreshing}
        loading={isRefreshing}>{m.status_action_refresh_errored_partitions()}</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
