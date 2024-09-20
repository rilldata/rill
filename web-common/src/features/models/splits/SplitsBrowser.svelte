<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import * as Select from "@rilldata/web-common/components/select";
  import Button from "../../../components/button/Button.svelte";
  import CollapsibleSectionTitle from "../../../layout/CollapsibleSectionTitle.svelte";
  import { V1Resource } from "../../../runtime-client";
  import SplitsTable from "./SplitsTable.svelte";

  export let resource: V1Resource;

  let active = true;

  let openFilterMenu = false;
  const options = [
    { value: "all", label: "all" },
    { value: "pending", label: "pending" },
    { value: "errors", label: "errors" },
  ];
  let selectedFilter = options[0];
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
          <div class="flex justify-end">
            <Select.Root
              items={options}
              onSelectedChange={(newSelection) => {
                if (!newSelection) return;
                selectedFilter = newSelection;
              }}
              bind:open={openFilterMenu}
            >
              <Select.Trigger
                class="outline-none border-none w-fit px-0 gap-x-0.5"
              >
                <Button type="text" label="Filter splits">
                  <span class="text-gray-700 hover:text-inherit">
                    Showing <b>{selectedFilter.label}</b>
                  </span>
                </Button>
              </Select.Trigger>
              <Select.Content sameWidth={false} align="end">
                {#each options as option (option.value)}
                  <Select.Item
                    value={option.value}
                    label={option.label}
                    class={`text-xs flex flex-col items-start ${
                      selectedFilter.value === option.value ? "font-bold" : ""
                    }`}
                  />
                {/each}
              </Select.Content>
            </Select.Root>
          </div>
          <SplitsTable
            {resource}
            whereErrored={selectedFilter.value === "errors"}
            wherePending={selectedFilter.value === "pending"}
          />
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
