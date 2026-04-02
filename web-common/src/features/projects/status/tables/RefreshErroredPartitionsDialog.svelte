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
        Refresh Errored Partitions for {modelName}?
      </AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This will re-execute all partitions that failed during their last run.
          The refresh will happen in the background.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        disabled={isRefreshing}
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button
        type="primary"
        onClick={handleRefresh}
        disabled={isRefreshing}
        loading={isRefreshing}>Refresh Errored Partitions</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
