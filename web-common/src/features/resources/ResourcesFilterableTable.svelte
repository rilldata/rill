<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import VirtualizedTable from "@rilldata/web-common/components/table/VirtualizedTable.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import ResourceTypeBadge from "@rilldata/web-common/features/entity-management/ResourceTypeBadge.svelte";
  import {
    ResourceKind,
    prettyResourceKind,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import RefreshAllSourcesAndModelsConfirmDialog from "@rilldata/web-common/features/resources/RefreshAllSourcesAndModelsConfirmDialog.svelte";
  import TagFilterDropdown from "@rilldata/web-common/features/resources/TagFilterDropdown.svelte";
  import {
    filterableTypes,
    filterResources,
    getAvailableTags,
    getStatusPriority,
    statusFilters,
  } from "@rilldata/web-common/features/resources/resource-filter-utils";
  import ActionsCell from "@rilldata/web-common/features/projects/status/ActionsCell.svelte";
  import NameCell from "@rilldata/web-common/features/projects/status/NameCell.svelte";
  import RefreshCell from "@rilldata/web-common/features/projects/status/RefreshCell.svelte";
  import ResourceErrorMessage from "@rilldata/web-common/features/projects/status/ResourceErrorMessage.svelte";
  import ResourceSpecDialog from "@rilldata/web-common/features/projects/status/ResourceSpecDialog.svelte";
  import {
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import type { ColumnDef } from "tanstack-table-8-svelte-5";
  import { renderComponent } from "tanstack-table-8-svelte-5";

  /** All resources (unfiltered). Filtering is handled internally. */
  export let resources: V1Resource[];
  export let isLoading = false;
  export let isError = false;
  export let errorMessage = "";
  /** Whether the "Refresh all" button should be disabled */
  export let isRefreshDisabled = false;
  /** Callback when user confirms "Refresh all sources and models" */
  export let onRefreshAll: () => void;
  /** Callback after a single-resource refresh completes (to refetch the list) */
  export let onRefetch: () => void;
  /** Pre-set status filters (e.g. from overview errors section) */
  export let selectedStatuses: string[] = [];
  /** Pre-set type filters (e.g. from overview resources section) */
  export let selectedTypes: string[] = [];
  /** Pre-set tag filters */
  export let selectedTags: string[] = [];
  /** Two-way bindable search text */
  export let searchText = "";
  /** Fixed table container height (web-admin uses 550) */
  export let containerHeight: number | undefined = undefined;
  /** Text shown when no resources match filters */
  export let emptyText: string | undefined = undefined;

  let isConfirmDialogOpen = false;
  let filterDropdownOpen = false;
  let statusDropdownOpen = false;
  let openDropdownResourceKey = "";

  // Describe dialog state
  let specDialogOpen = false;
  let specResourceName = "";
  let specResourceKind = "";
  let specResource: V1Resource | undefined = undefined;

  function handleDescribe(name: string, kind: string, resource: V1Resource) {
    specResourceName = name;
    specResourceKind = kind;
    specResource = resource;
    specDialogOpen = true;
  }

  const setDropdownOpen = (resourceKey: string, isOpen: boolean) => {
    openDropdownResourceKey = isOpen ? resourceKey : "";
  };

  const isDropdownOpen = (resourceKey: string) => {
    return openDropdownResourceKey === resourceKey;
  };

  $: availableTags = getAvailableTags(resources);

  $: filteredResources = filterResources(
    resources,
    selectedTypes,
    searchText,
    selectedStatuses,
    selectedTags,
  );

  $: tableData = filteredResources.filter(
    (resource) =>
      resource.meta?.name?.kind !== ResourceKind.Component &&
      resource.meta?.name?.kind !== ResourceKind.ProjectParser,
  );

  $: hasActiveFilters =
    selectedTypes.length > 0 ||
    searchText ||
    selectedStatuses.length > 0 ||
    selectedTags.length > 0;

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
    selectedTags = [];
    searchText = "";
  }

  const columns: ColumnDef<V1Resource, any>[] = [
    {
      accessorKey: "title",
      header: "Type",
      accessorFn: (row) => row.meta?.name?.kind,
      cell: ({ row }) =>
        renderComponent(ResourceTypeBadge, {
          kind: row.original.meta?.name?.kind as ResourceKind,
        }),
    },
    {
      accessorFn: (row) => row.meta?.name?.name,
      header: "Name",
      cell: ({ getValue }) =>
        renderComponent(NameCell, {
          name: getValue() as string,
        }),
    },
    {
      accessorFn: (row) => row.meta?.reconcileStatus,
      header: "Status",
      sortingFn: (rowA, rowB) =>
        getStatusPriority(rowB.original.meta?.reconcileStatus) -
        getStatusPriority(rowA.original.meta?.reconcileStatus),
      cell: ({ row }) =>
        renderComponent(ResourceErrorMessage, {
          message: row.original.meta?.reconcileError ?? "",
          status:
            row.original.meta?.reconcileStatus ??
            V1ReconcileStatus.RECONCILE_STATUS_UNSPECIFIED,
        }),
      meta: {
        marginLeft: "1",
      },
    },
    {
      accessorFn: (row) => row.meta?.stateUpdatedOn,
      header: "Last refresh",
      sortDescFirst: true,
      cell: (info) =>
        renderComponent(RefreshCell, {
          date: info.getValue() as string,
        }),
    },
    {
      accessorFn: (row) => row.meta?.reconcileOn,
      header: "Next refresh",
      cell: (info) =>
        renderComponent(RefreshCell, {
          date: info.getValue() as string,
        }),
    },
    {
      accessorKey: "actions",
      header: "",
      cell: ({ row }) => {
        const status = row.original.meta?.reconcileStatus;
        const isRowReconciling =
          status === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
          status === V1ReconcileStatus.RECONCILE_STATUS_RUNNING;
        const resourceKey = `${row.original.meta?.name?.kind}:${row.original.meta?.name?.name}`;
        return renderComponent(ActionsCell, {
          resourceKind: row.original.meta?.name?.kind ?? "",
          resourceName: row.original.meta?.name?.name ?? "",
          canRefresh:
            !isRowReconciling &&
            (row.original.meta?.name?.kind === ResourceKind.Model ||
              row.original.meta?.name?.kind === ResourceKind.Source),
          resource: row.original,
          onRefresh: onRefetch,
          onDescribe: handleDescribe,
          isDropdownOpen: isDropdownOpen(resourceKey),
          onDropdownOpenChange: (isOpen: boolean) =>
            setDropdownOpen(resourceKey, isOpen),
        });
      },
      enableSorting: false,
      meta: {
        widthPercent: 0,
      },
    },
  ];
</script>

<section class="flex flex-col gap-y-4 size-full">
  <div class="flex items-center justify-between">
    <h2 class="text-lg font-medium">Resources</h2>
  </div>

  <!-- Search, Filter, and Action Controls -->
  <div class="flex flex-row items-center gap-x-4 min-h-9">
    <div class="flex-1 min-w-0 min-h-9">
      <Search
        bind:value={searchText}
        placeholder="Search"
        large
        autofocus={false}
        showBorderOnFocus={false}
        retainValueOnMount
      />
    </div>

    <DropdownMenu.Root bind:open={filterDropdownOpen}>
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {filterDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium">
          {#if selectedTypes.length === 0}
            All types
          {:else if selectedTypes.length === 1}
            {prettyResourceKind(selectedTypes[0])}
          {:else}
            {prettyResourceKind(selectedTypes[0])}, +{selectedTypes.length - 1} other{selectedTypes.length >
            2
              ? "s"
              : ""}
          {/if}
        </span>
        {#if filterDropdownOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
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
      <DropdownMenu.Trigger
        class="min-w-fit min-h-9 flex flex-row gap-1 items-center rounded-sm border bg-input {statusDropdownOpen
          ? 'bg-gray-200'
          : 'hover:bg-surface-hover'} px-2 py-1"
      >
        <span class="text-fg-secondary font-medium">
          {#if selectedStatuses.length === 0}
            All statuses
          {:else if selectedStatuses.length === 1}
            {statusFilters.find((s) => s.value === selectedStatuses[0])
              ?.label ?? selectedStatuses[0]}
          {:else}
            {statusFilters.find((s) => s.value === selectedStatuses[0])?.label},
            +{selectedStatuses.length - 1} other{selectedStatuses.length > 2
              ? "s"
              : ""}
          {/if}
        </span>
        {#if statusDropdownOpen}
          <CaretUpIcon size="12px" />
        {:else}
          <CaretDownIcon size="12px" />
        {/if}
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

    <TagFilterDropdown tags={availableTags} bind:selectedTags />

    {#if hasActiveFilters}
      <button
        class="shrink-0 text-sm text-primary-500 hover:text-primary-600 whitespace-nowrap"
        onclick={clearFilters}
      >
        Clear
      </button>
    {/if}

    <Button
      type="secondary"
      large
      class="shrink-0 whitespace-nowrap"
      onClick={() => {
        isConfirmDialogOpen = true;
      }}
      disabled={isRefreshDisabled}
    >
      <span class="hidden lg:inline">Refresh all sources and models</span>
      <span class="lg:hidden">Refresh all</span>
    </Button>
  </div>

  {#if isLoading && resources.length === 0}
    <DelayedSpinner isLoading={true} size="16px" />
  {:else if isError}
    <div class="text-red-500">
      Error loading resources: {errorMessage}
    </div>
  {:else}
    <VirtualizedTable
      data={tableData}
      {columns}
      columnLayout="minmax(95px, 108px) minmax(100px, 3fr) 48px minmax(80px, 2fr) minmax(100px, 2fr) 56px"
      {containerHeight}
      {emptyText}
    />
  {/if}

  <slot name="after-table" />
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={onRefreshAll}
/>

<ResourceSpecDialog
  bind:open={specDialogOpen}
  resourceName={specResourceName}
  resourceKind={specResourceKind}
  resource={specResource}
/>
