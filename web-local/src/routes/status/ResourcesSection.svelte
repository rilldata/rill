<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import RefreshAllSourcesAndModelsConfirmDialog from "@rilldata/web-common/features/resources/RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import {
    filterableTypes,
    filterResources,
    statusFilters,
  } from "@rilldata/web-common/features/resources/resource-filter-utils";
  import {
    createRuntimeServiceCreateTrigger,
    createRuntimeServiceListResources,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import ResourcesTable from "./ResourcesTable.svelte";

  /** Pre-set status filters when navigating from the overview errors section */
  export let initialStatusFilter: string[] = [];
  /** Pre-set type filters when navigating from the overview resources section */
  export let initialTypeFilter: string[] = [];

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let filterDropdownOpen = false;
  let statusDropdownOpen = false;
  let searchText = "";
  let selectedTypes: string[] = initialTypeFilter;
  let selectedStatuses: string[] = initialStatusFilter;

  // React to prop changes (e.g., clicking errors section switches tab and sets filter)
  $: selectedStatuses = initialStatusFilter;
  $: selectedTypes = initialTypeFilter;

  $: resourcesQuery = createRuntimeServiceListResources(
    $runtime.instanceId,
    {},
    { query: { refetchInterval: 5000 } },
  );

  $: resources = $resourcesQuery.data?.resources ?? [];
  $: isLoading = $resourcesQuery.isLoading;
  $: isError = $resourcesQuery.isError;
  $: error = $resourcesQuery.error;

  $: filteredResources = filterResources(
    resources,
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

  $: hasActiveFilters =
    selectedTypes.length > 0 || searchText || selectedStatuses.length > 0;

  $: hasReconcilingSourcesOrModels = resources.some(
    (r) =>
      (r.meta?.name?.kind === ResourceKind.Source ||
        r.meta?.name?.kind === ResourceKind.Model) &&
      (r.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
        r.meta?.reconcileStatus === V1ReconcileStatus.RECONCILE_STATUS_RUNNING),
  );

  function refreshAllSourcesAndModels() {
    void $createTrigger
      .mutateAsync({
        instanceId: $runtime.instanceId,
        data: { all: true },
      })
      .then(() => {
        void queryClient.invalidateQueries({
          queryKey: getRuntimeServiceListResourcesQueryKey(
            $runtime.instanceId,
            undefined,
          ),
        });
      });
  }
</script>

<section class="flex flex-col gap-y-4 size-full">
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

    {#if hasActiveFilters}
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
      />
    </div>

    <Button
      type="secondary"
      onClick={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={hasReconcilingSourcesOrModels}
    >
      Refresh all sources and models
    </Button>
  </div>

  {#if isLoading && resources.length === 0}
    <DelayedSpinner isLoading={true} size="16px" />
  {:else if isError}
    <div class="text-red-500">
      Error loading resources: {error?.message}
    </div>
  {:else}
    <ResourcesTable
      data={filteredResources}
      onRefresh={() => $resourcesQuery.refetch()}
    />
  {/if}
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
