<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import type { Selected } from "bits-ui";
  import CollapsibleSectionTitle from "../../../layout/CollapsibleSectionTitle.svelte";
  import type { V1Resource } from "../../../runtime-client";
  import SplitsFilter from "./SplitsFilter.svelte";
  import SplitsTable from "./SplitsTable.svelte";

  export let resource: V1Resource;

  let active = true;

  let selectedFilter = "all";

  function onFilterChange(newSelection: Selected<string>) {
    selectedFilter = newSelection.value;
  }
</script>

<section>
  <CollapsibleSectionTitle tooltipText="model splits" bind:active>
    Model splits
  </CollapsibleSectionTitle>

  {#if active}
    <div class="line-wrapper">
      {#if resource.model?.state?.splitsHaveErrors}
        <CancelCircle size="12" className="text-red-600" />
      {/if}
      <span class="help-text">
        {resource.model?.state?.splitsHaveErrors
          ? "Some splits have errors.  "
          : "All splits were successful.   "}
        <Dialog.Root>
          <Dialog.Trigger class="text-primary-500 font-medium">
            View splits
          </Dialog.Trigger>
          <Dialog.Content class="max-w-screen-xl">
            <Dialog.Header>
              <Dialog.Title>Model splits</Dialog.Title>
            </Dialog.Header>
            <div class="flex justify-end">
              <SplitsFilter {selectedFilter} onChange={onFilterChange} />
            </div>
            <SplitsTable
              {resource}
              whereErrored={selectedFilter === "errors"}
              wherePending={selectedFilter === "pending"}
            />
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
    @apply text-xs text-gray-500;
  }
</style>
