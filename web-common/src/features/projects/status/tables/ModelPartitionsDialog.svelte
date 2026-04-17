<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Search } from "@rilldata/web-common/components/search";
  import PartitionsTable from "@rilldata/web-common/features/models/partitions/PartitionsTable.svelte";
  import PartitionsFilter from "@rilldata/web-common/features/models/partitions/PartitionsFilter.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import {
    shouldFilterByErrored,
    shouldFilterByPending,
    type PartitionFilterType,
  } from "./utils";

  export let open = false;
  export let resource: V1Resource | null = null;
  export let onClose: () => void = () => {};

  let selectedFilter: PartitionFilterType = "all";
  let searchText = "";

  function onFilterChange(value: string) {
    selectedFilter = value as PartitionFilterType;
  }

  $: modelName = resource?.meta?.name?.name ?? "";
  $: whereErrored = shouldFilterByErrored(selectedFilter);
  $: wherePending = shouldFilterByPending(selectedFilter);
</script>

<Dialog.Root
  {open}
  onOpenChange={(o) => {
    if (!o) {
      selectedFilter = "all";
      searchText = "";
      onClose();
    }
  }}
>
  <Dialog.Content class="max-w-screen-xl h-[40vh] flex flex-col gap-y-4">
    <Dialog.Header>
      <Dialog.Title>Model Partitions: {modelName}</Dialog.Title>
    </Dialog.Header>

    {#if resource}
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
        <PartitionsFilter {selectedFilter} onChange={onFilterChange} />
      </div>
      <div class="flex-1 min-h-0 overflow-auto">
        <PartitionsTable
          {resource}
          {whereErrored}
          {wherePending}
          {searchText}
        />
      </div>
    {/if}
  </Dialog.Content>
</Dialog.Root>
