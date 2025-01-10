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
  export let onRefresh: () => void;

  async function handleRefresh() {
    try {
      onRefresh();
      open = false;
    } catch (error) {
      console.error("Failed to refresh resource:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <!-- <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger> -->
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Refresh all sources and models?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This will refresh all sources and models in the project.
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
      <Button type="primary" on:click={handleRefresh}>Yes, refresh</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
