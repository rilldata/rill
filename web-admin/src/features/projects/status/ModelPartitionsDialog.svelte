<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import PartitionsTable from "@rilldata/web-common/features/models/partitions/PartitionsTable.svelte";
  import PartitionsFilter from "@rilldata/web-common/features/models/partitions/PartitionsFilter.svelte";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import type { Selected } from "bits-ui";

  export let open = false;
  export let resource: V1Resource | null = null;
  export let onClose: () => void = () => {};

  let selectedFilter = "all";

  function onFilterChange(newSelection: Selected<string>) {
    if (newSelection?.value) {
      selectedFilter = newSelection.value;
    }
  }

  $: modelName = resource?.meta?.name?.name ?? "";
  $: whereErrored = selectedFilter === "errors";
  $: wherePending = selectedFilter === "pending";
</script>

<Dialog.Root
  {open}
  onOpenChange={(o) => {
    if (!o) {
      selectedFilter = "all";
      onClose();
    }
  }}
>
  <Dialog.Content class="max-w-screen-xl max-h-[90vh] flex flex-col">
    <Dialog.Header>
      <Dialog.Title>Model Partitions: {modelName}</Dialog.Title>
    </Dialog.Header>

    {#if resource}
      <div class="flex justify-end mb-4">
        <PartitionsFilter {selectedFilter} onChange={onFilterChange} />
      </div>
      <div class="flex-1 overflow-hidden">
        <PartitionsTable {resource} {whereErrored} {wherePending} />
      </div>
    {/if}
  </Dialog.Content>
</Dialog.Root>
