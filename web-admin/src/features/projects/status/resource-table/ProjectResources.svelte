<script lang="ts">
  import { page } from "$app/stores";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceGetResource,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { SingletonProjectParserName } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ProjectResourcesTable from "./ProjectResourcesTable.svelte";
  import RefreshAllSourcesAndModelsConfirmDialog from "./RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import { useResources } from "../selectors";
  import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store";
  import { filterResources } from "./utils";
  import {
    createUrlFilterSync,
    parseArrayParam,
    parseStringParam,
  } from "../url-filter-sync";
  import { onMount } from "svelte";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  const filterSync = createUrlFilterSync([
    { key: "kind", type: "array" },
    { key: "status", type: "array" },
    { key: "q", type: "string" },
  ]);
  filterSync.init($page.url);

  let isConfirmDialogOpen = false;
  let filterDropdownOpen = false;
  let statusDropdownOpen = false;
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
    filterSync.syncToUrl($page.url, {
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

  $: ({ instanceId } = $runtime);

  $: resources = useResources(instanceId);

  // Parse errors
  $: projectParserQuery = createRuntimeServiceGetResource(
    instanceId,
    {
      "name.kind": ResourceKind.ProjectParser,
      "name.name": SingletonProjectParserName,
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
    void $createTrigger
      .mutateAsync({
        instanceId,
        data: { all: true },
      })
      .then(() => {
        void queryClient.invalidateQueries({
          queryKey: getRuntimeServiceListResourcesQueryKey(
            instanceId,
            undefined,
          ),
        });
      });
  }
</script>

<section class="flex flex-col gap-y-4">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
  </div>

  <!-- Filter and Search Controls -->
  <div class="flex items-center gap-x-3">
    <DropdownMenu.Root bind:open={filterDropdownOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <Button builders={[builder]} type="tertiary">
          <span class="flex items-center gap-x-1.5">
            {#if selectedTypes.length === 0}
              All types
            {:else if selectedTypes.length === 1}
              {prettyResourceKind(selectedTypes[0])}
            {:else}
              {prettyResourceKind(selectedTypes[0])}, +{selectedTypes.length -
                1} other{selectedTypes.length > 2 ? "s" : ""}
            {/if}
            {#if filterDropdownOpen}
              <CaretUpIcon size="12px" />
            {:else}
              <CaretDownIcon size="12px" />
            {/if}
          </span>
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-48">
        {#each filterableTypes as type}
          <DropdownMenu.CheckboxItem
            checked={selectedTypes.includes(type)}
            onCheckedChange={() => toggleType(type)}
          >
            {prettyResourceKind(type)}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    <DropdownMenu.Root bind:open={statusDropdownOpen}>
      <DropdownMenu.Trigger asChild let:builder>
        <Button builders={[builder]} type="tertiary">
          <span class="flex items-center gap-x-1.5">
            {#if selectedStatuses.length === 0}
              All statuses
            {:else if selectedStatuses.length === 1}
              {statusFilters.find((s) => s.value === selectedStatuses[0])
                ?.label ?? selectedStatuses[0]}
            {:else}
              {statusFilters.find((s) => s.value === selectedStatuses[0])
                ?.label}, +{selectedStatuses.length - 1} other{selectedStatuses.length >
              2
                ? "s"
                : ""}
            {/if}
            {#if statusDropdownOpen}
              <CaretUpIcon size="12px" />
            {:else}
              <CaretDownIcon size="12px" />
            {/if}
          </span>
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="start" class="w-48">
        {#each statusFilters as status}
          <DropdownMenu.CheckboxItem
            checked={selectedStatuses.includes(status.value)}
            onCheckedChange={() => toggleStatus(status.value)}
          >
            {status.label}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    {#if selectedTypes.length > 0 || searchText || selectedStatuses.length > 0}
      <button
        class="text-sm text-primary-500 hover:text-primary-600"
        on:click={clearFilters}
      >
        Clear filters
      </button>
    {/if}

    <!-- Spacer -->
    <div class="flex-1" />

    <div class="w-64">
      <Search
        bind:value={searchText}
        placeholder="Search by name..."
        autofocus={false}
        retainValueOnMount={true}
      />
    </div>

    <Button
      type="secondary"
      onClick={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={isRefreshButtonDisabled}
    >
      Refresh all sources and models
    </Button>
  </div>

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
