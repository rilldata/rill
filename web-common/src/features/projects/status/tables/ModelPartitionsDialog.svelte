<script lang="ts">
  import { onDestroy } from "svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { Search } from "@rilldata/web-common/components/search";
  import PartitionsTable from "@rilldata/web-common/features/models/partitions/PartitionsTable.svelte";
  import PartitionsFilter from "@rilldata/web-common/features/models/partitions/PartitionsFilter.svelte";
  import BackfillPartitionsPanel from "@rilldata/web-common/features/models/partitions/BackfillPartitionsPanel.svelte";
  import {
    getRuntimeServiceGetModelPartitionsQueryKey,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    shouldFilterByErrored,
    shouldFilterByPending,
    type PartitionFilterType,
  } from "./utils";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let open = false;
  export let resource: V1Resource | null = null;
  export let onClose: () => void = () => {};

  let selectedFilter: PartitionFilterType = "all";
  let searchText = "";
  let showBackfill = false;
  let pollInterval: ReturnType<typeof setInterval> | null = null;

  const runtimeClient = useRuntimeClient();
  const queryClient = useQueryClient();

  function onFilterChange(value: string) {
    selectedFilter = value as PartitionFilterType;
  }

  function stopPolling() {
    if (pollInterval) {
      clearInterval(pollInterval);
      pollInterval = null;
    }
  }

  // Partitions execute asynchronously after a backfill is triggered,
  // so poll every 2 seconds for up to 30 seconds, like TriggerPartition.
  function onBackfillTriggered() {
    showBackfill = false;
    stopPolling();
    const queryKey = getRuntimeServiceGetModelPartitionsQueryKey(
      runtimeClient.instanceId,
      { model: modelName },
    );
    void queryClient.invalidateQueries({ queryKey });
    let pollCount = 0;
    pollInterval = setInterval(() => {
      pollCount++;
      void queryClient.invalidateQueries({ queryKey });
      if (pollCount >= 15) stopPolling();
    }, 2000);
  }

  onDestroy(stopPolling);

  $: modelName = resource?.meta?.name?.name ?? "";
  $: isIncremental = !!resource?.model?.spec?.incremental;
  $: whereErrored = shouldFilterByErrored(selectedFilter);
  $: wherePending = shouldFilterByPending(selectedFilter);
</script>

<Dialog.Root
  {open}
  onOpenChange={(o) => {
    if (!o) {
      selectedFilter = "all";
      searchText = "";
      showBackfill = false;
      stopPolling();
      onClose();
    }
  }}
>
  <Dialog.Content
    class="max-w-screen-xl {showBackfill
      ? 'h-[75vh]'
      : 'h-[40vh]'} flex flex-col gap-y-4"
  >
    <Dialog.Header>
      <Dialog.Title>
        {showBackfill
          ? m.status_backfill_title()
          : m.status_model_partitions()}: {modelName}
      </Dialog.Title>
    </Dialog.Header>

    {#if resource && showBackfill}
      <BackfillPartitionsPanel
        {resource}
        onDone={() => {
          showBackfill = false;
        }}
        onTriggered={onBackfillTriggered}
      />
    {:else if resource}
      <div class="flex flex-row items-center gap-x-4 min-h-9">
        <div class="flex-1 min-w-0 min-h-9">
          <Search
            bind:value={searchText}
            large
            autofocus={false}
            showBorderOnFocus={false}
            retainValueOnMount
          />
        </div>
        <PartitionsFilter {selectedFilter} onChange={onFilterChange} />
        {#if isIncremental}
          <Button
            type="secondary"
            onClick={() => {
              showBackfill = true;
            }}
          >
            {m.status_backfill_button()}
          </Button>
        {/if}
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
