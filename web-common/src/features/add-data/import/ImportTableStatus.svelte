<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import {
    ImportTableMode,
    type ImportTableRunner,
  } from "@rilldata/web-common/features/add-data/import/ImportTableRunner.ts";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { CheckIcon, XIcon } from "lucide-svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";

  export let runner: ImportTableRunner;
  export let onBack: () => void;

  const { mode, error, details, currentFilePath } = runner;

  const Steps = [
    {
      mode: ImportTableMode.CreateModel,
      pendingLabel: "Ingesting data...",
      doneLabel: "Ingested data",
      failedLabel: "Ingesting data failed.",
    },
    {
      mode: ImportTableMode.CreateMetrics,
      pendingLabel: "Creating Metrics View...",
      doneLabel: "Created Metrics View",
      failedLabel: "Creating Metrics View failed.",
    },
    {
      mode: ImportTableMode.CreateExplore,
      pendingLabel: "Generating Explore dashboard...",
      doneLabel: "Generated Explore dashboard",
      failedLabel: "Generating Explore dashboard failed.",
    },
  ];

  $: hasErrored = !!$error;
  $: currentFileRoute = $currentFilePath
    ? `/files${addLeadingSlash($currentFilePath)}`
    : "/";
</script>

<div class="flex flex-col gap-4 p-6 mx-auto w-fit">
  <div class="flex justify-center">
    {#if hasErrored}
      <AlertCircleOutline size="32px" className="text-destructive" />
    {:else}
      <Spinner status={EntityStatus.Running} size="32px" />
    {/if}
  </div>
  <div class="text-center">Creating your dashboard</div>
  {#each Steps as step (step.mode)}
    <div class="flex flex-row items-center gap-4 text-fg-tertiary">
      {#if $mode > step.mode}
        <CheckIcon size="14px" />
        <div>{step.doneLabel}</div>
      {:else if hasErrored}
        {#if $mode === step.mode}
          <AlertCircleOutline size="14px" className="text-destructive" />
          <div>{step.failedLabel}</div>
        {:else}
          <XIcon size="14px" className="text-destructive" />
          <div>{step.pendingLabel}</div>
        {/if}
      {:else}
        <LoadingSpinner size="14px" />
        <div>{step.pendingLabel}</div>
      {/if}
    </div>
  {/each}
</div>
{#if $error}
  <div class="w-96 mx-auto mb-4 text-destructive">
    <div class="text-sm mb-2">{$error}</div>
    {#if $details}<div>{$details}</div>{/if}
  </div>
{/if}
<div class="flex flex-row items-center gap-2 mb-4 mx-auto">
  {#if hasErrored}
    <Button type="secondary" noStroke href={currentFileRoute} onClick={onBack}>
      Back
    </Button>
  {/if}
  <Button type="secondary" href={currentFileRoute}>
    Skip and view project
  </Button>
  {#if hasErrored}
    <Button type="primary">Retry (TODO)</Button>
  {/if}
</div>
