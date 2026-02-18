<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import GraphContainer from "@rilldata/web-common/features/resource-graph/navigation/GraphContainer.svelte";
  import {
    parseGraphUrlParams,
    urlParamsToSeeds,
    type KindToken,
  } from "@rilldata/web-common/features/resource-graph/navigation/seed-parser";
  import type { ResourceStatusFilterValue } from "@rilldata/web-common/features/resource-graph/shared/types";
  import WorkspaceContainer from "@rilldata/web-common/layout/workspace/WorkspaceContainer.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Search from "@rilldata/web-common/components/search/Search.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import {
    createRuntimeServiceCreateTrigger,
    getRuntimeServiceListResourcesQueryKey,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  const queryClient = useQueryClient();
  const createTrigger = createRuntimeServiceCreateTrigger();

  $: ({ instanceId } = $runtime);

  // Parse URL parameters using new API (kind/resource instead of seed)
  $: urlParams = parseGraphUrlParams($page.url);
  $: seeds = urlParamsToSeeds(urlParams);
  $: activeKind = urlParams.kind;

  // Node type filter config
  type NodeTypeOption = { label: string; token: KindToken };
  const nodeTypeOptions: NodeTypeOption[] = [
    { label: "Source Models", token: "sources" },
    { label: "Models", token: "models" },
    { label: "Metric Views", token: "metrics" },
    { label: "Dashboards", token: "dashboards" },
  ];

  let nodeTypeDropdownOpen = false;

  function selectNodeType(token: KindToken | null) {
    nodeTypeDropdownOpen = false;
    if (token) {
      goto(`/graph?kind=${token}`);
    } else {
      goto("/graph");
    }
  }

  // Search and filter state
  let searchQuery = "";
  let selectedStatuses: ResourceStatusFilterValue[] = [];
  let statusDropdownOpen = false;

  type StatusOption = { label: string; value: ResourceStatusFilterValue };
  const statusOptions: StatusOption[] = [
    { label: "OK", value: "ok" },
    { label: "Pending", value: "pending" },
    { label: "Errored", value: "errored" },
  ];

  function toggleStatus(value: ResourceStatusFilterValue) {
    if (selectedStatuses.includes(value)) {
      selectedStatuses = selectedStatuses.filter((s) => s !== value);
    } else {
      selectedStatuses = [...selectedStatuses, value];
    }
  }

  function clearFilters() {
    searchQuery = "";
    selectedStatuses = [];
    if (activeKind) {
      goto("/graph");
    }
  }

  $: hasActiveFilters =
    searchQuery || selectedStatuses.length > 0 || !!activeKind;

  let isConfirmDialogOpen = false;

  function refreshAllSourcesAndModels() {
    isConfirmDialogOpen = false;
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
      })
      .catch((err) => {
        console.error("Failed to refresh all sources and models:", err);
      });
  }
</script>

<svelte:head>
  <title>Rill Developer | Project graph</title>
</svelte:head>

<WorkspaceContainer inspector={false}>
  <div slot="header" class="header">
    <div class="header-row">
      <div class="header-left">
        <h1>Project graph</h1>
      </div>
    </div>
    <p>Visualize dependencies between sources, models, dashboards, and more.</p>

    <!-- Filter bar matching cloud status resource format -->
    <div class="filter-bar">
      <!-- Node type filter -->
      <DropdownMenu.Root bind:open={nodeTypeDropdownOpen}>
        <DropdownMenu.Trigger asChild let:builder>
          <Button builders={[builder]} type="tertiary">
            <span class="flex items-center gap-x-1.5">
              {#if activeKind}
                {nodeTypeOptions.find((o) => o.token === activeKind)?.label ??
                  activeKind}
              {:else}
                All types
              {/if}
              {#if nodeTypeDropdownOpen}
                <CaretUpIcon size="12px" />
              {:else}
                <CaretDownIcon size="12px" />
              {/if}
            </span>
          </Button>
        </DropdownMenu.Trigger>
        <DropdownMenu.Content align="start" class="w-48">
          <DropdownMenu.Item on:click={() => selectNodeType(null)}>
            All types
          </DropdownMenu.Item>
          <DropdownMenu.Separator />
          {#each nodeTypeOptions as option}
            <DropdownMenu.Item
              on:click={() => selectNodeType(option.token)}
              class={activeKind === option.token ? "font-semibold" : ""}
            >
              {option.label}
            </DropdownMenu.Item>
          {/each}
        </DropdownMenu.Content>
      </DropdownMenu.Root>

      <!-- Status filter -->
      <DropdownMenu.Root bind:open={statusDropdownOpen}>
        <DropdownMenu.Trigger asChild let:builder>
          <Button builders={[builder]} type="tertiary">
            <span class="flex items-center gap-x-1.5">
              {#if selectedStatuses.length === 0}
                All statuses
              {:else if selectedStatuses.length === 1}
                {statusOptions.find((s) => s.value === selectedStatuses[0])
                  ?.label ?? selectedStatuses[0]}
              {:else}
                {statusOptions.find((s) => s.value === selectedStatuses[0])
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
          {#each statusOptions as status}
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
          bind:value={searchQuery}
          placeholder="Search by node..."
          autofocus={false}
        />
      </div>

      <Button
        type="secondary"
        onClick={() => {
          isConfirmDialogOpen = true;
        }}
      >
        Refresh all sources and models
      </Button>
    </div>
  </div>

  <div slot="body" class="graph-wrapper">
    <GraphContainer
      {seeds}
      {searchQuery}
      statusFilter={selectedStatuses}
      showSummary={false}
    />
  </div>
</WorkspaceContainer>

<AlertDialog.Root bind:open={isConfirmDialogOpen}>
  <AlertDialog.Content>
    <AlertDialog.Header>
      <AlertDialog.Title>Refresh all sources and models?</AlertDialog.Title>
      <AlertDialog.Description>
        <div class="flex flex-col gap-y-2 mt-1">
          <p>This will refresh all project sources and models.</p>
          <p>
            <span class="font-medium">Note:</span> To refresh a single resource, click
            the '...' button on a node and select the refresh option.
          </p>
        </div>
      </AlertDialog.Description>
    </AlertDialog.Header>
    <AlertDialog.Footer>
      <Button
        type="tertiary"
        onClick={() => {
          isConfirmDialogOpen = false;
        }}>Cancel</Button
      >
      <Button type="primary" onClick={refreshAllSourcesAndModels}
        >Yes, refresh</Button
      >
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<style lang="postcss">
  .header {
    @apply px-4 pt-3 pb-2;
  }

  .header h1 {
    @apply text-lg font-semibold text-fg-primary;
  }

  .header-row {
    @apply flex items-center justify-between;
  }

  .header p {
    @apply text-sm text-fg-secondary mt-1;
  }

  .filter-bar {
    @apply flex items-center gap-x-3 mt-3;
  }

  .graph-wrapper {
    @apply h-full w-full;
  }
</style>
