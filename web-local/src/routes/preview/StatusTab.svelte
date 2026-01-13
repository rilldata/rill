<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import RefreshIcon from "@rilldata/web-common/components/icons/RefreshIcon.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  let resources: any[] = [];
  let loading = false;
  let isRequesting = false;
  let error: string | null = null;
  let resourceStats = {
    total: 0,
    succeeded: 0,
    failed: 0,
    compiling: 0,
  };
  let runtimeStatus = "healthy";

  async function loadStatus() {
    // Prevent concurrent requests
    if (isRequesting) return;

    try {
      isRequesting = true;
      loading = true;
      error = null;

      if (!$runtime?.instanceId || !$runtime?.host) {
        error = "Waiting for runtime to initialize...";
        loading = false;
        isRequesting = false;
        return;
      }

      // Fetch the list of resources from the runtime
      const response = await fetch(
        `${$runtime.host}/v1/instances/${$runtime.instanceId}/resources`,
      );

      if (!response.ok) {
        throw new Error(`Failed to fetch resources: ${response.statusText}`);
      }

      const data = await response.json();
      resources = data?.resources || [];

      // Calculate stats
      resourceStats.total = resources.length;
      resourceStats.succeeded = resources.filter(
        (r: any) => r.state === "READY" || r.resourceState === "ready",
      ).length;
      resourceStats.failed = resources.filter(
        (r: any) =>
          r.state === "ERROR" ||
          r.resourceState === "error" ||
          r.error !== undefined,
      ).length;
      resourceStats.compiling = resources.filter(
        (r: any) =>
          r.state === "RECONCILING" ||
          r.resourceState === "reconciling" ||
          r.resourceState === "compiling",
      ).length;

      // Determine overall status
      if (resourceStats.failed > 0) {
        runtimeStatus = "unhealthy";
      } else if (resourceStats.compiling > 0) {
        runtimeStatus = "compiling";
      } else {
        runtimeStatus = "healthy";
      }
    } catch (err) {
      error =
        err instanceof Error ? err.message : "Failed to load status";
      console.error("Error loading status:", err);
    } finally {
      loading = false;
      isRequesting = false;
    }
  }

  onMount(() => {
    loadStatus();
    // Refresh every 10 seconds to avoid overwhelming the runtime
    const interval = setInterval(loadStatus, 10000);
    return () => clearInterval(interval);
  });

  // Retry when runtime becomes available
  $: if ($runtime?.instanceId && $runtime?.host && error?.includes("Waiting")) {
    loadStatus();
  }

  function getStatusColor(status: string) {
    switch (status) {
      case "healthy":
        return "text-green-600 dark:text-green-400 bg-green-50 dark:bg-green-900";
      case "compiling":
        return "text-yellow-600 dark:text-yellow-400 bg-yellow-50 dark:bg-yellow-900";
      case "unhealthy":
        return "text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900";
      default:
        return "";
    }
  }

  function getResourceStatusColor(state: string) {
    switch (state?.toUpperCase()) {
      case "READY":
        return "text-green-600 dark:text-green-400";
      case "ERROR":
        return "text-red-600 dark:text-red-400";
      case "RECONCILING":
      case "COMPILING":
        return "text-yellow-600 dark:text-yellow-400";
      default:
        return "text-gray-600 dark:text-gray-400";
    }
  }

  function getStatusBadge(state: string) {
    switch (state?.toUpperCase()) {
      case "READY":
        return "✓";
      case "ERROR":
        return "✕";
      case "RECONCILING":
      case "COMPILING":
        return "⟳";
      default:
        return "−";
    }
  }

  function getResourceState(resource: any): string {
    return resource.state || resource.resourceState || "unknown";
  }
</script>

<div class="h-full w-full overflow-auto bg-white dark:bg-gray-950">
  <div class="p-8 space-y-8 max-w-7xl">
    <!-- Runtime Health Section -->
    <section>
      <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
        Runtime Status
      </h2>

      <div class={`p-6 rounded-lg ${getStatusColor(runtimeStatus)}`}>
        <div class="flex items-center gap-4">
          <div class="text-4xl">
            {#if runtimeStatus === "healthy"}
              ✓
            {:else if runtimeStatus === "compiling"}
              ⟳
            {:else}
              ✕
            {/if}
          </div>
          <div>
            <p class="font-semibold text-lg capitalize">{runtimeStatus}</p>
            <p class="text-sm opacity-75">
              {#if runtimeStatus === "healthy"}
                All resources are compiled and ready
              {:else if runtimeStatus === "compiling"}
                {resourceStats.compiling} resource{resourceStats.compiling !== 1 ? "s" : ""} compiling
              {:else}
                {resourceStats.failed} resource{resourceStats.failed !== 1 ? "s" : ""} have errors
              {/if}
            </p>
          </div>
        </div>
      </div>
    </section>

    <!-- Statistics Section -->
    <section>
      <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-4">
        Resource Statistics
      </h2>

      <div class="grid grid-cols-4 gap-4">
        <div class="p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800">
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-1">Total</p>
          <p class="text-3xl font-bold text-gray-900 dark:text-white">
            {resourceStats.total}
          </p>
        </div>
        <div class="p-4 bg-green-50 dark:bg-green-900 rounded-lg border border-green-200 dark:border-green-800">
          <p class="text-sm text-green-600 dark:text-green-400 mb-1">Succeeded</p>
          <p class="text-3xl font-bold text-green-600 dark:text-green-400">
            {resourceStats.succeeded}
          </p>
        </div>
        <div class="p-4 bg-yellow-50 dark:bg-yellow-900 rounded-lg border border-yellow-200 dark:border-yellow-800">
          <p class="text-sm text-yellow-600 dark:text-yellow-400 mb-1">Compiling</p>
          <p class="text-3xl font-bold text-yellow-600 dark:text-yellow-400">
            {resourceStats.compiling}
          </p>
        </div>
        <div class="p-4 bg-red-50 dark:bg-red-900 rounded-lg border border-red-200 dark:border-red-800">
          <p class="text-sm text-red-600 dark:text-red-400 mb-1">Failed</p>
          <p class="text-3xl font-bold text-red-600 dark:text-red-400">
            {resourceStats.failed}
          </p>
        </div>
      </div>
    </section>

    <!-- Resource List Section -->
    <section>
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">
          Resources
        </h2>
        <Button
          type="secondary"
          onClick={loadStatus}
          disabled={loading}
          square
          label="Refresh resources"
        >
          <RefreshIcon size="14px" />
        </Button>
      </div>

      {#if error}
        <div class="bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 rounded p-4 mb-6">
          <p class="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      {/if}

      <div class="space-y-2 max-h-96 overflow-y-auto">
        {#if resources.length === 0}
          <p class="text-gray-500 dark:text-gray-400 text-sm py-4">
            {loading ? "Loading resources..." : "No resources found"}
          </p>
        {:else}
          {#each resources as resource (resource.name)}
            <div
              class="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-800 rounded hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors"
            >
              <div class="flex items-center gap-3 flex-1 min-w-0">
                <ResourceTypeBadge kind={resource.resource_kind || resource.kind} />
                <div class="min-w-0 flex-1">
                  <p class="font-medium text-sm truncate text-gray-900 dark:text-white">
                    {resource.name}
                  </p>
                  <p class="text-xs text-gray-500 dark:text-gray-500">
                    {resource.kind || resource.resource_kind || "unknown"}
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-2">
                <span
                  class={`text-xs px-3 py-1 rounded capitalize ${getResourceStatusColor(getResourceState(resource))} opacity-75`}
                >
                  {getResourceState(resource).toLowerCase()}
                </span>
              </div>
            </div>
          {/each}
        {/if}
      </div>
    </section>
  </div>
</div>
