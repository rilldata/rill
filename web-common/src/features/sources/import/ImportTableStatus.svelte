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

  const Steps = [
    {
      mode: ImportTableMode.CreateModel,
      pendingLabel: "Ingesting data...",
      doneLabel: "Ingested data",
    },
    {
      mode: ImportTableMode.CreateMetrics,
      pendingLabel: "Creating Metrics View...",
      doneLabel: "Created Metrics View",
    },
    {
      mode: ImportTableMode.CreateExplore,
      pendingLabel: "Generating Explore dashboard...",
      doneLabel: "Generated Explore dashboard",
    },
  ];
</script>

<div class="flex flex-col gap-4 p-6 mx-auto w-fit">
  <div class="flex justify-center">
    <Spinner status={EntityStatus.Running} size="32px" />
  </div>
  <div class="text-center">Creating your dashboard</div>
  {#each Steps as step (step.mode)}
    <div class="status">
      {#if $mode > step.mode}
        <CheckIcon size="14px" />
        <div>{step.doneLabel}</div>
      {:else}
        <LoadingSpinner size="14px" />
        <div>{step.pendingLabel}</div>
      {/if}
    </div>
  {/each}
</div>

<style lang="postcss">
  .status {
    @apply flex flex-row items-center gap-4;
  }
</style>
