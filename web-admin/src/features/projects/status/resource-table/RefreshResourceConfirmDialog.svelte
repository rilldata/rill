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
  export let name: string;
  export let onRefresh: () => void;
  export let refreshType: "full" | "incremental" = "full";

  const MAX_NAME_LENGTH = 37;

  function truncateName(str: string): string {
    if (str.length > MAX_NAME_LENGTH) {
      return str.substring(0, MAX_NAME_LENGTH) + "...";
    }
    return str;
  }

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
      <AlertDialogTitle>
        {refreshType === "full" ? "Full Refresh" : "Incremental Refresh"}
        <span class="font-semibold" title={name}>{truncateName(name)}</span>?
      </AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {#if refreshType === "full"}
            ⚠️ Warning: A full refresh will re-ingest ALL data from scratch.
            This operation can take a significant amount of time and will update
            all dependent resources. Only proceed if you're certain this is
            necessary.
          {:else}
            Refreshing this resource will update all dependent resources.
          {/if}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" onClick={handleRefresh}>Yes, refresh</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
