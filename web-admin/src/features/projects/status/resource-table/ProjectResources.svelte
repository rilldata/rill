<script lang="ts">
  import { page } from "$app/stores";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTriggerMutation,
    createRuntimeServiceGetResource,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { SingletonProjectParserName } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { TableToolbar } from "@rilldata/web-common/components/table-toolbar";
  import type { FilterGroup } from "@rilldata/web-common/components/table-toolbar/types";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "@rilldata/web-common/features/resources/RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { useResources } from "../selectors";
  import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store";
  import { filterResources } from "@rilldata/web-common/features/resources/resource-filter-utils";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "@rilldata/web-common/lib/url-filter-sync";
  import { onMount } from "svelte";

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();
  const createTrigger =
    createRuntimeServiceCreateTriggerMutation(runtimeClient);

  const filterSync = createUrlFilterSync([
    { key: "kind", type: "array" },
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  filterSync.init($page.url);

  let isConfirmDialogOpen = false;
  let searchText = parseStringParam($page.url.searchParams.get("q"));
  let selectedTypes = parseArrayParam($page.url.searchParams.get("kind"));
  let selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
  let mounted = false;

  // Sync URL → local state on external navigation (back/forward)
  $: if (mounted && filterSync.hasExternalNavigation($page.url)) {
    filterSync.markSynced($page.url);
    selectedTypes = parseArrayParam($page.url.searchParams.get("kind"));
    selectedStatuses = parseArrayParam($page.url.searchParams.get("status"));
    searchText = parseStringParam($page.url.searchParams.get("q"));
  }

  // Sync filter state → URL
  $: if (mounted) {
    filterSync.syncToUrl({
      kind: selectedTypes,
      status: selectedStatuses,
      q: searchText,
    });
  }

  onMount(() => {
    mounted = true;
  });

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "Error", value: "error" },
    { label: "Warn", value: "warn" },
    { label: "OK", value: "ok" },
  ];

  // Resource types available for filtering (excluding internal types)
  const filterableTypes = [
    ResourceKind.Source,
    ResourceKind.Model,
    ResourceKind.MetricsView,
    ResourceKind.Explore,
    ResourceKind.Canvas,
    ResourceKind.Theme,
    ResourceKind.Report,
    ResourceKind.Alert,
    ResourceKind.API,
    ResourceKind.Connector,
  ];

  $: filterGroups = [
    {
      label: "Type",
      key: "kind",
      options: filterableTypes.map((t) => ({
        value: t,
        label: prettyResourceKind(t),
      })),
      selected: selectedTypes,
      defaultValue: [],
      multiSelect: true,
    },
    {
      label: "Status",
      key: "status",
      options: statusFilters.map((s) => ({
        value: s.value,
        label: s.label,
      })),
      selected: selectedStatuses,
      defaultValue: [],
      multiSelect: true,
    },
  ] satisfies FilterGroup[];

  $: resources = useResources(runtimeClient);

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    runtimeClient,
    {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  $: parseErrors =
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [];

  $: hasReconcilingResources = $resources.data?.resources?.some(
    isResourceReconciling,
  );

  $: isRefreshButtonDisabled = hasReconcilingResources;

  // Filter resources by type, search text, and status
  $: filteredResources = filterResources(
    $resources.data?.resources,
    selectedTypes,
    searchText,
    selectedStatuses,
  );

  function toggleType(type: string) {
    if (selectedTypes.includes(type)) {
      selectedTypes = selectedTypes.filter((t) => t !== type);
    } else {
      selectedTypes = [...selectedTypes, type];
    }
  }

  function toggleStatus(status: string) {
    if (selectedStatuses.includes(status)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== status);
    } else {
      selectedStatuses = [...selectedStatuses, status];
    }
  }

  function clearFilters() {
    selectedTypes = [];
    selectedStatuses = [];
    searchText = "";
  }

  function refreshAllSourcesAndModels() {
    void $createTrigger.mutateAsync({ all: true }).then(() => {
      void queryClient.invalidateQueries({
        queryKey: getRuntimeServiceListResourcesQueryKey(
          runtimeClient.instanceId,
          undefined,
        ),
      });
    });
  }
</script>

<section class="flex flex-col gap-y-4">
  <h2 class="text-lg font-medium">Resources</h2>

  <TableToolbar
    {searchText}
    onSearchChange={(text) => {
      searchText = text;
    }}
    {filterGroups}
    onFilterChange={(key, value) => {
      if (key === "kind") toggleType(value);
      if (key === "status") toggleStatus(value);
    }}
    onClearAllFilters={clearFilters}
    showSort={false}
  >
    <Button
      type="secondary"
      large
      class="shrink-0 whitespace-nowrap"
      onClick={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={isRefreshButtonDisabled}
    >
      <span class="hidden lg:inline">Refresh all sources and models</span>
      <span class="lg:hidden">Refresh all</span>
    </Button>
  </TableToolbar>

  {#if $resources.isLoading}
    <DelayedSpinner isLoading={true} size="16px" />
  {:else if $resources.isError}
    <div class="text-red-500">
      Error loading resources: {$resources.error?.message}
    </div>
  {:else if $resources.data}
    <ProjectResourcesTable data={filteredResources} />
  {/if}

  <div class="parse-errors">
    <h3 class="parse-errors-header">
      Parse Errors
      {#if parseErrors.length > 0}
        <span class="parse-errors-badge">{parseErrors.length}</span>
      {/if}
    </h3>
    {#if parseErrors.length === 0}
      <p class="text-sm text-fg-secondary">No parse errors</p>
    {:else}
      <div class="parse-errors-list">
        {#each parseErrors as error ((error.filePath ?? "") + ":" + error.message)}
          <div class="parse-error-item">
            {#if error.filePath}
              <span class="parse-error-file">{error.filePath}</span>
            {/if}
            <span class="parse-error-message">{error.message}</span>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>

<style lang="postcss">
  .parse-errors {
    @apply pt-4 mt-2;
  }
  .parse-errors-header {
    @apply text-sm font-semibold text-fg-primary flex items-center gap-2 mb-3;
  }
  .parse-errors-badge {
    @apply text-xs font-semibold text-white bg-red-500 rounded-full px-1.5 py-0.5 min-w-[20px] text-center;
  }
  .parse-errors-list {
    @apply flex flex-col gap-2;
  }
  .parse-error-item {
    @apply flex flex-col gap-0.5 px-3 py-2 rounded-md bg-red-50 text-sm;
  }
  .parse-error-file {
    @apply font-mono text-xs text-fg-secondary;
  }
  .parse-error-message {
    @apply text-red-700;
  }
</style>
