<script lang="ts">
  import { goto } from "$app/navigation";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { XIcon } from "lucide-svelte";
  import AlertCircleOutline from "@rilldata/web-common/components/icons/AlertCircleOutline.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    type ImportAddDataStep,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import {
    cleanupImportStep,
    runImportSteps,
  } from "@rilldata/web-common/features/add-data/manager/steps/import.ts";
  import { onMount } from "svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import FeatherCheckCircle from "@rilldata/web-common/components/icons/FeatherCheckCircle.svelte";

  export let importAddDataStep: ImportAddDataStep;
  export let onBack: () => void;
  export let onClose: () => void;

  const runtimeClient = useRuntimeClient();
  const initialAddDataStep = { ...importAddDataStep };

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
      step: ImportDataStep.CreateDashboard,
      pendingLabel: "Generating dashboard...",
      doneLabel: "Generated dashboard",
      failedLabel: "Generating dashboard failed.",
    },
  ];
  const steps = importAddDataStep.config.importSteps.map(
    (s) => StepLabels.find((label) => label.step === s)!,
  );

  let importStep = ImportDataStep.Init;
  $: currentFileRoute = "/";
  let error: string | null = null;
  $: hasErrored = !!error;

  async function runImport() {
    importStep = ImportDataStep.Init;
    error = null;

    try {
      await runImportSteps(
        runtimeClient,
        importAddDataStep.config,
        (step, currentFilePath) => {
          importStep = step;
          if (currentFilePath) {
            currentFileRoute = `/files${addLeadingSlash(currentFilePath)}`;
          }
        },
      );
      onClose();
      return goto(currentFileRoute);
    } catch (e) {
      error = e?.response?.data?.message ?? e?.message ?? null;
    }
  }

  function rerunImport() {
    importAddDataStep = { ...initialAddDataStep };
    return runImport();
  }

  async function cleanupAndBack() {
    await cleanupImportStep(
      runtimeClient,
      queryClient,
      importAddDataStep.config,
    );

    onBack();
  }

  onMount(runImport);
</script>

<div class="flex flex-col gap-4 p-6 mx-auto w-fit justify-center">
  <div class="flex justify-center">
    {#if hasErrored}
      <AlertCircleOutline size="16px" className="text-destructive" />
    {:else}
      <Spinner status={EntityStatus.Running} size="30px" />
    {/if}
  </div>
  <div class="flex flex-col gap-y-2">
    <div class="text-center font-semibold text-[18px]">
      Creating your dashboard
    </div>
    <div class="flex flex-col gap-y-1">
      {#each steps as s (s.step)}
        <div class="flex flex-row items-center gap-2 text-fg-tertiary text-sm">
          {#if importStep > s.step}
            <FeatherCheckCircle size="18px" />
            <div>{s.doneLabel}</div>
          {:else if hasErrored}
            {#if importStep === s.step}
              <AlertCircleOutline size="18px" className="text-destructive" />
              <div>{s.failedLabel}</div>
            {:else}
              <XIcon size="18px" className="text-destructive" />
              <div>{s.pendingLabel}</div>
            {/if}
          {:else}
            <LoadingSpinner size="18px" />
            <div>{s.pendingLabel}</div>
          {/if}
        </div>
      {/each}
    </div>
  </div>

  {#if error}
    <div class="w-96 mx-auto text-destructive">
      <div class="text-sm mb-2">{error}</div>
    </div>
  {/if}

  <div class="flex flex-row items-center gap-2 py-6 mx-auto">
    {#if hasErrored}
      <Button type="secondary" noStroke onClick={cleanupAndBack}>Back</Button>
    {/if}
    <Button type="tertiary" href={currentFileRoute} onClick={onClose} large>
      Skip and view project
    </Button>
    {#if hasErrored}
      <Button type="primary" onClick={rerunImport}>Retry</Button>
    {/if}
  </div>
</div>
