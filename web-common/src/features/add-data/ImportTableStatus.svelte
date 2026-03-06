<script lang="ts">
  import { goto } from "$app/navigation";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { CheckIcon, XIcon } from "lucide-svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    type AddDataConfig,
    type ImportAddDataStep,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { runImportStep } from "@rilldata/web-common/features/add-data/steps/import.ts";
  import { onMount } from "svelte";

  export let config: AddDataConfig;
  export let importAddDataStep: ImportAddDataStep;
  export let onBack: () => void;

  $: ({ importStep } = importAddDataStep);

  const Steps = [
    {
      step: ImportDataStep.CreateModel,
      pendingLabel: "Ingesting data...",
      doneLabel: "Ingested data",
      failedLabel: "Ingesting data failed.",
    },
    {
      step: ImportDataStep.CreateMetricsView,
      pendingLabel: "Creating Metrics View...",
      doneLabel: "Created Metrics View",
      failedLabel: "Creating Metrics View failed.",
    },
    {
      step: ImportDataStep.CreateExplore,
      pendingLabel: "Generating Explore dashboard...",
      doneLabel: "Generated Explore dashboard",
      failedLabel: "Generating Explore dashboard failed.",
    },
  ];

  let currentFileRoute: string = "/";
  let error: string | null = null;
  $: hasErrored = !!error;

  async function runImport() {
    try {
      while (importAddDataStep.importStep.step !== ImportDataStep.Done) {
        importAddDataStep = await runImportStep(
          config,
          importAddDataStep,
          (newRoute) => (currentFileRoute = newRoute),
        );
      }
      return goto(currentFileRoute);
    } catch (e) {
      error = e?.response?.data?.message ?? e?.message ?? null;
    }
  }

  onMount(runImport);
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
  {#each Steps as s (s.step)}
    <div class="flex flex-row items-center gap-4 text-fg-tertiary">
      {#if importStep.step > s.step}
        <CheckIcon size="14px" />
        <div>{s.doneLabel}</div>
      {:else if hasErrored}
        {#if importStep.step > s.step}
          <AlertCircleOutline size="14px" className="text-destructive" />
          <div>{s.failedLabel}</div>
        {:else}
          <XIcon size="14px" className="text-destructive" />
          <div>{s.pendingLabel}</div>
        {/if}
      {:else}
        <LoadingSpinner size="14px" />
        <div>{s.pendingLabel}</div>
      {/if}
    </div>
  {/each}
</div>
{#if $error}
  <div class="w-96 mx-auto mb-4 text-destructive">
    <div class="text-sm mb-2">{$error}</div>
  </div>
{/if}
<div class="flex flex-row items-center gap-2 mb-4 mx-auto">
  {#if hasErrored}
    <Button type="secondary" noStroke onClick={onBack}>Back</Button>
  {/if}
  <Button type="secondary" href={currentFileRoute}>
    Skip and view project
  </Button>
  {#if hasErrored}
    <Button type="primary">Retry (TODO)</Button>
  {/if}
</div>
