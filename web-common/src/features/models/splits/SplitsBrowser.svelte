<script lang="ts">
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import CollapsibleSectionTitle from "../../../layout/CollapsibleSectionTitle.svelte";
  import {
    V1Resource,
    createRuntimeServiceGetModelSplits,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import SplitsTable from "./SplitsTable.svelte";

  export let resource: V1Resource;

  let active = true;

  $: modelName = resource?.meta?.name?.name as string;

  $: splitsQuery = createRuntimeServiceGetModelSplits(
    $runtime.instanceId,
    modelName,
  );
  $: ({ data } = $splitsQuery);
</script>

<CollapsibleSectionTitle tooltipText="model splits" bind:active>
  Model splits
</CollapsibleSectionTitle>

{#if active}
  <div class="wrapper">
    <span class="help-text">
      {resource.model?.state?.splitsHaveErrors
        ? "Some splits have errors.  "
        : "All splits were successful.   "}
      {#if true}
        <Dialog.Root>
          <Dialog.Trigger class="text-primary-500 font-medium">
            View splits
          </Dialog.Trigger>
          <Dialog.Content class="max-w-screen-xl">
            <Dialog.Header>
              <Dialog.Title>Model splits</Dialog.Title>
            </Dialog.Header>
            <div class="max-h-[80vh]">
              <SplitsTable {resource} splits={data.splits} />
            </div>
          </Dialog.Content>
        </Dialog.Root>
      {/if}
    </span>
  </div>
{/if}

<style lang="postcss">
  .wrapper {
    @apply flex flex-col gap-y-2 items-start;
  }

  .help-text {
    @apply text-xs text-gray-500;
  }
</style>
