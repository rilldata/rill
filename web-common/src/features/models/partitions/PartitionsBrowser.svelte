<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import CollapsibleSectionTitle from "../../../layout/CollapsibleSectionTitle.svelte";
  import type { V1Resource } from "../../../runtime-client";
  import PartitionsFilter from "./PartitionsFilter.svelte";
  import PartitionsTable from "./PartitionsTable.svelte";

  export let resource: V1Resource;

  let active = true;
  let open = false;
  let selectedFilter = "all";
  let searchText = "";

  function onFilterChange(newValue: string) {
    selectedFilter = newValue;
  }

  function onOpenChange(value: boolean) {
    open = value;
    if (!value) {
      selectedFilter = "all";
      searchText = "";
    }
  }
</script>

<section>
  <CollapsibleSectionTitle tooltipText="model partitions" bind:active>
    Model partitions
  </CollapsibleSectionTitle>

  {#if active}
    <div class="line-wrapper">
      {#if resource.model?.state?.partitionsHaveErrors}
        <CancelCircle size="12" className="text-red-600" />
      {/if}
      <span class="help-text">
        {resource.model?.state?.partitionsHaveErrors
          ? "Some partitions have errors.  "
          : "All partitions were successful.   "}
        <Dialog.Root {open} {onOpenChange}>
          <Dialog.Trigger class="text-primary-500 font-medium">
            View partitions
          </Dialog.Trigger>
          <Dialog.Content class="max-w-screen-xl max-h-[90vh] flex flex-col">
            <Dialog.Header>
              <Dialog.Title>Model partitions</Dialog.Title>
            </Dialog.Header>
            <div class="flex items-center gap-x-3 mb-4">
              <div class="w-64">
                <Search
                  bind:value={searchText}
                  placeholder="Search partitions"
                  autofocus={false}
                />
              </div>
              <div class="ml-auto">
                <PartitionsFilter {selectedFilter} onChange={onFilterChange} />
              </div>
            </div>
            <div class="flex-1 min-h-0 overflow-auto">
              <PartitionsTable
                {resource}
                whereErrored={selectedFilter === "errors"}
                wherePending={selectedFilter === "pending"}
                {searchText}
              />
            </div>
          </Dialog.Content>
        </Dialog.Root>
      </span>
    </div>
  {/if}
</section>

<style lang="postcss">
  section {
    @apply px-4 flex flex-col gap-y-2;
  }

  .line-wrapper {
    @apply flex items-center gap-x-1;
  }

  .help-text {
    @apply text-xs text-fg-secondary;
  }
</style>
