<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";
  import ResourcesTable from "./ResourcesTable.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  let resources: V1Resource[] = [];
  let isLoading = false;
  let isError = false;
  let error: Error | null = null;
  let hasReconcilingResources = false;

  async function loadResources() {
    try {
      isError = false;
      error = null;
      isLoading = true;

      if (!$runtime?.instanceId || !$runtime?.host) {
        throw new Error("Runtime not initialized");
      }

      const response = await fetch(
        `${$runtime.host}/v1/instances/${$runtime.instanceId}/resources`,
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch resources: ${response.statusText}`);
      }

      const data = await response.json();
      resources = data?.resources || [];

      // Check if any resources are reconciling
      hasReconcilingResources = resources.some((r) => {
        const status = r.meta?.reconcileStatus;
        return (
          status === 2 || // RECONCILE_STATUS_PENDING
          status === 3 // RECONCILE_STATUS_RUNNING
        );
      });
    } catch (err) {
      isError = true;
      error = err instanceof Error ? err : new Error("Unknown error");
      console.error("Error loading resources:", err);
    } finally {
      isLoading = false;
    }
  }

  onMount(() => {
    loadResources();
    // Refresh every 5 seconds
    const interval = setInterval(loadResources, 5000);
    return () => clearInterval(interval);
  });

</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
  </div>

  {#if isLoading && resources.length === 0}
    <DelayedSpinner isLoading={true} size="16px" />
  {:else if isError}
    <div class="text-red-500">
      Error loading resources: {error?.message}
    </div>
  {:else}
    <ResourcesTable data={resources} onRefresh={loadResources} />
  {/if}
</section>
