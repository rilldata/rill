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
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";

  export let open = false;
  export let name: string;
  export let resourceKind: string;
  export let onRefresh: () => Promise<void> | void;
  export let refreshType: "full" | "incremental" = "full";

  let isLoading = false;
  let error: string | null = null;

  async function handleRefresh() {
    try {
      isLoading = true;
      error = null;

      if (!$runtime?.instanceId || !$runtime?.host) {
        throw new Error("Runtime not initialized");
      }

      const requestBody = resourceKind === ResourceKind.Model
        ? {
            models: [
              {
                model: name,
                full: refreshType === "full",
              },
            ],
          }
        : {
            resources: [
              {
                kind: resourceKind,
                name: name,
              },
            ],
          };

      const url = `${$runtime.host}/v1/instances/${$runtime.instanceId}/trigger`;
      console.log("Triggering refresh for", { resourceKind, name, refreshType, url, requestBody });

      const response = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(requestBody),
      });

      console.log("Refresh response:", { status: response.status, statusText: response.statusText });

      if (!response.ok) {
        const errorText = await response.text();
        console.error("Refresh error response:", errorText);
        throw new Error(`Failed to trigger refresh: ${response.statusText}`);
      }

      console.log("Refresh triggered successfully, reloading resources...");
      await onRefresh();
      open = false;
    } catch (err) {
      error = err instanceof Error ? err.message : "Unknown error";
      console.error("Failed to refresh resource:", err);
    } finally {
      isLoading = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>
        {refreshType === "full" ? "Full Refresh" : "Incremental Refresh"}
        {name}?
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
        {#if error}
          <div class="mt-3 text-red-500 text-sm">{error}</div>
        {/if}
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}
        disabled={isLoading}>Cancel</Button
      >
      <Button type="primary" on:click={handleRefresh} disabled={isLoading}>
        {isLoading ? "Refreshing..." : "Yes, refresh"}
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
