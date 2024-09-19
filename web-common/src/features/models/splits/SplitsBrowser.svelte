<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import SelectMenu from "../../../components/menu/shadcn/SelectMenu.svelte";
  import CollapsibleSectionTitle from "../../../layout/CollapsibleSectionTitle.svelte";
  import { V1Resource } from "../../../runtime-client";
  import SplitsTable from "./SplitsTable.svelte";

  export let resource: V1Resource;

  let active = true;

  let selectedFilter: "all" | "errors" | "pending" = "errors";
  let options = [
    { main: "all", key: "all" },
    { main: "pending", key: "pending" },
    { main: "errors", key: "errors" },
  ];
</script>

<CollapsibleSectionTitle tooltipText="model splits" bind:active>
  Model splits
</CollapsibleSectionTitle>

{#if active}
  <div class="wrapper">
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
          <div class="flex justify-end gap-x-2">
            <SelectMenu
              fixedText="Showing"
              {options}
              selections={[selectedFilter]}
              on:select={(event) => (selectedFilter = event.detail.key)}
              ariaLabel="Filter splits"
            />
            <!-- <Select.Root bind:value={selectedFilter}>
                <Select.Trigger class="w-[200px]">
                  <Select.Value placeholder="Select abc" />
                </Select.Trigger>
                <Select.Content>
                  <Select.Item value="all">Show all</Select.Item>
                  <Select.Item value="pending">Show pending</Select.Item>
                  <Select.Item value="errors">Show errored</Select.Item>
                </Select.Content>
              </Select.Root> -->
          </div>
          <div class="max-h-[80vh]">
            <SplitsTable
              {resource}
              whereErrored={selectedFilter === "errors"}
              wherePending={selectedFilter === "pending"}
            />
          </div>
        </Dialog.Content>
      </Dialog.Root>
    </span>
  </div>
{/if}

<style lang="postcss">
  .wrapper {
    @apply flex items-center gap-x-1;
  }

  .help-text {
    @apply text-xs text-gray-500;
  }
</style>
