<script lang="ts">
  import CollapsibleSectionTitle from "@rilldata/web-common/layout/CollapsibleSectionTitle.svelte";
  import type { V1Resource } from "../../../runtime-client";
  import IncrementalStateTable from "./IncrementalStateTable.svelte";

  export let resource: V1Resource;

  let active = true;

  $: hasIncrementalState = !!resource.model?.state?.incrementalState;
</script>

<section>
  <CollapsibleSectionTitle tooltipText="incremental state" bind:active>
    Incremental processing
  </CollapsibleSectionTitle>

  {#if active}
    {#if !hasIncrementalState}
      <div class="help-text">Each refresh will append new records.</div>
    {:else if hasIncrementalState}
      <div class="wrapper">
        <span class="help-text">
          Use this state to determine which records to process on each model
          run.
        </span>
        <IncrementalStateTable {resource} />
      </div>
    {/if}
  {/if}
</section>

<style lang="postcss">
  section {
    @apply px-4 flex flex-col gap-y-2;
  }

  .wrapper {
    @apply flex flex-col gap-y-2 items-start;
  }

  .help-text {
    @apply text-xs text-gray-500;
  }
</style>
