<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";
  import Button from "web-common/src/components/button/Button.svelte";
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
  import { useResources } from "./selectors";
  import { isResourceReconciling } from "@rilldata/web-admin/lib/refetch-interval-store";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  let isConfirmDialogOpen = false;
  let filterDropdownOpen = false;
  let searchText = "";
  let selectedTypes: string[] = [];

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

  // Filter resources by type and search text
  $: filteredResources = filterResources(
    $resources.data?.resources,
    selectedTypes,
    searchText,
  );

  function filterResources(
    resources: V1Resource[] | undefined,
    types: string[],
    search: string,
  ): V1Resource[] {
    if (!resources) return [];

    return resources.filter((r) => {
      const kind = r.meta?.name?.kind;
      const name = r.meta?.name?.name ?? "";

      const matchesType = types.length === 0 || types.includes(kind ?? "");
      const matchesSearch =
        !search || name.toLowerCase().includes(search.toLowerCase());

      return matchesType && matchesSearch;
    });
  }

  function toggleType(type: string) {
    if (selectedTypes.includes(type)) {
      selectedTypes = selectedTypes.filter((t) => t !== type);
    } else {
      selectedTypes = [...selectedTypes, type];
    }
  }

  function clearFilters() {
    selectedTypes = [];
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
              {prettyResourceKind(selectedTypes[0])}, +{selectedTypes.length - 1} other{selectedTypes.length > 2 ? "s" : ""}
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
            onSelect={(e) => e.preventDefault()}
          >
            {prettyResourceKind(type)}
          </DropdownMenu.CheckboxItem>
        {/each}
      </DropdownMenu.Content>
    </DropdownMenu.Root>

    {#if selectedTypes.length > 0 || searchText}
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
</section>

<RefreshAllSourcesAndModelsConfirmDialog
  bind:open={isConfirmDialogOpen}
  onRefresh={refreshAllSourcesAndModels}
/>
