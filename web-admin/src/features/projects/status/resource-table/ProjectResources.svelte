<script lang="ts">
  import { page } from "$app/stores";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
    V1ReconcileStatus,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
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

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  // Initialize filter from URL search params (e.g. ?kind=rill.runtime.v1.Model or ?error=true)
  const kindParam = $page.url.searchParams.get("kind");
  const errorParam = $page.url.searchParams.get("error") === "true";

  let isConfirmDialogOpen = false;
  let filterDropdownOpen = false;
  let statusDropdownOpen = false;
  let showDagViewer = false;
  let searchText = "";
  let selectedTypes: string[] = kindParam ? [kindParam] : [];
  let selectedStatuses: string[] = errorParam ? ["error"] : [];

  type StatusFilter = { label: string; value: string };
  const statusFilters: StatusFilter[] = [
    { label: "Error", value: "error" },
    { label: "Warn", value: "warn" },
    { label: "OK", value: "ok" },
  ];

  // Resource types available for filtering (excluding internal types)
  const filterableTypes = [
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

  function getResourceStatus(r: V1Resource): string {
    if (r.meta?.reconcileError) return "error";
    const status = r.meta?.reconcileStatus;
    if (
      status === V1ReconcileStatus.RECONCILE_STATUS_PENDING ||
      status === V1ReconcileStatus.RECONCILE_STATUS_RUNNING
    )
      return "warn";
    return "ok";
  }

  function filterResources(
    resources: V1Resource[] | undefined,
    types: string[],
    search: string,
    statuses: string[],
  ): V1Resource[] {
    if (!resources) return [];

    return resources.filter((r) => {
      const kind = r.meta?.name?.kind;
      const name = r.meta?.name?.name ?? "";

      const matchesType = types.length === 0 || types.includes(kind ?? "");
      const matchesSearch =
        !search || name.toLowerCase().includes(search.toLowerCase());
      const matchesStatus =
        statuses.length === 0 || statuses.includes(getResourceStatus(r));

      return matchesType && matchesSearch && matchesStatus;
    });
  }

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
    <label class="flex items-center gap-2 cursor-pointer">
      <span class="text-sm text-fg-secondary">DAG Viewer</span>
      <Switch bind:checked={showDagViewer} small />
    </label>
  </div>

  {#if showDagViewer}
    <div
      class="flex items-center justify-center h-64 border border-border rounded-lg text-fg-secondary text-sm"
    >
      DAG Viewer coming soon
    </div>
  {:else}
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
      <DelayedSpinner isLoading={$resources.isLoading} size="16px" />
    {:else if $resources.isError}
      <div class="text-red-500">
        Error loading resources: {$resources.error?.message}
      </div>
    {:else if $resources.data}
      <ProjectResourcesTable data={filteredResources} />
    {/if}
  {/if}
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
