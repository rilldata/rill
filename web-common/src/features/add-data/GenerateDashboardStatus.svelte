<script lang="ts">
  import { goto } from "$app/navigation";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { CheckIcon, XIcon } from "lucide-svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    type ImportAddDataStep,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { runImportStep } from "@rilldata/web-common/features/add-data/steps/import.ts";
  import { onMount } from "svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";

  export let importAddDataStep: ImportAddDataStep;
  export let onBack: () => void;
  export let onClose: () => void;

  const runtimeClient = useRuntimeClient();
  const initialAddDataStep = { ...importAddDataStep };

  $: ({ importStep } = importAddDataStep);

  const StepLabels = [
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
      pendingLabel: "Generating explore dashboard...",
      doneLabel: "Generated explore dashboard",
      failedLabel: "Generating explore dashboard failed.",
    },
    {
      step: ImportDataStep.CreateCanvas,
      pendingLabel: "Generating canvas dashboard...",
      doneLabel: "Generated canvas dashboard",
      failedLabel: "Generating canvas dashboard failed.",
    },
  ];
  const steps = importAddDataStep.config.importSteps.map(
    (s) => StepLabels.find((label) => label.step === s)!,
  );

  $: currentFileRoute = importAddDataStep.currentFilePath
    ? `/files/${addLeadingSlash(importAddDataStep.currentFilePath)}`
    : "/";
  let error: string | null = null;
  $: hasErrored = !!error;

  async function runImport() {
    try {
      while (importAddDataStep.importStep !== ImportDataStep.Done) {
        importAddDataStep = await runImportStep(
          runtimeClient,
          importAddDataStep,
        );
      }
      onClose();
      if (!importAddDataStep.currentFilePath) return goto("/");
      return goto(
        `/files/${addLeadingSlash(importAddDataStep.currentFilePath)}`,
      );
    } catch (e) {
      error = e?.response?.data?.message ?? e?.message ?? null;
    }
  }

  function rerunImport() {
    importAddDataStep = { ...initialAddDataStep };
    error = null;
    return runImport();
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
  {#each steps as s (s.step)}
    <div class="flex flex-row items-center gap-4 text-fg-tertiary">
      {#if importStep > s.step}
        <CheckIcon size="14px" />
        <div>{s.doneLabel}</div>
      {:else if hasErrored}
        {#if importStep > s.step}
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
{#if error}
  <div class="w-96 mx-auto mb-4 text-destructive">
    <div class="text-sm mb-2">{error}</div>
  </div>
{/if}
<div class="flex flex-row items-center gap-2 mb-4 mx-auto">
  {#if hasErrored}
    <Button type="secondary" noStroke onClick={onBack}>Back</Button>
  {/if}
  <Button type="secondary" href={currentFileRoute} onClick={onClose}>
    Skip and view project
  </Button>
  {#if hasErrored}
    <Button type="primary" onClick={rerunImport}>Retry</Button>
  {/if}
</div>
