<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import {
    ImportTableMode,
    type ImportTableRunner,
  } from "@rilldata/web-common/features/sources/import/ImportTableRunner.ts";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { CheckIcon } from "lucide-svelte";

  export let runner: ImportTableRunner;
  const { mode } = runner;
</script>

<div class="flex flex-col gap-4">
  <Spinner status={EntityStatus.Running} size="32px" />
  <div>Creating your dashboard</div>
  <div class="status">
    {#if $mode > ImportTableMode.CreateModel}
      <CheckIcon size="14px" />
      <div>Ingested data</div>
    {:else}
      <LoadingSpinner size="14px" />
      <div>Ingesting data...</div>
    {/if}
  </div>
  <div class="status">
    {#if $mode > ImportTableMode.CreateMetrics}
      <CheckIcon size="14px" />
      <div>Created Metrics View</div>
    {:else}
      <LoadingSpinner size="14px" />
      <div>Creating Metrics View...</div>
    {/if}
  </div>
  <div class="status">
    {#if $mode > ImportTableMode.CreateExplore}
      <CheckIcon size="14px" />
      <div>Generated Explore dashboard</div>
    {:else}
      <LoadingSpinner size="14px" />
      <div>Generating Explore dashboard...</div>
    {/if}
  </div>
</div>

<style lang="postcss">
  .status {
    @apply flex flex-row items-center gap-4;
  }
</style>
