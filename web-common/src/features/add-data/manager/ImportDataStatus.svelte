<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    type ImportAddDataStep,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { runImportStep } from "@rilldata/web-common/features/add-data/steps/import.ts";
  import { onMount } from "svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    WandIcon,
    CheckCircle2Icon,
    AlertCircleIcon,
    Loader2Icon,
  } from "lucide-svelte";
  import {
    createCanvasDashboardFromTableWithAgent,
    useCreateMetricsViewWithCanvasAndExploreUIAction,
  } from "@rilldata/web-common/features/metrics-views/ai-generation/generateMetricsView.ts";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes.ts";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import { addLeadingSlash } from "@rilldata/web-common/features/entity-management/entity-mappers.ts";

  export let importAddDataStep: ImportAddDataStep;
  export let onClose: () => void;

  const { ai, developerChat } = featureFlags;

  const runtimeClient = useRuntimeClient();

  $: currentFileRoute = importAddDataStep.currentFilePath
    ? `/files${addLeadingSlash(importAddDataStep.currentFilePath)}`
    : "/";
  $: sourceName = importAddDataStep.config.importTo.modelName ?? "";
  $: isDone = importAddDataStep.importStep === ImportDataStep.Done;
  let error: string | null = null;

  $: createDashboardFromTable =
    useCreateMetricsViewWithCanvasAndExploreUIAction(
      runtimeClient,
      importAddDataStep.config.connector,
      "",
      "",
      sourceName,
      BehaviourEventMedium.Button,
      MetricsEventSpace.Modal,
    );

  async function runImport() {
    try {
      while (importAddDataStep.importStep !== ImportDataStep.Done) {
        importAddDataStep = await runImportStep(
          runtimeClient,
          importAddDataStep,
        );
      }
      if (!importAddDataStep.currentFilePath) return goto("/");
      return goto(
        `/files${addLeadingSlash(importAddDataStep.currentFilePath)}`,
      );
    } catch (e) {
      error = e?.response?.data?.message ?? e?.message ?? null;
    }
  }

  async function generateMetrics() {
    onClose();
    if ($developerChat && $ai) {
      await createCanvasDashboardFromTableWithAgent(
        runtimeClient,
        importAddDataStep.config.connector,
        "",
        "",
        sourceName,
      );
    } else {
      await createDashboardFromTable();
    }
  }

  onMount(runImport);
</script>

<div class="flex flex-col gap-4 p-6 mx-auto w-full">
  {#if error}
    <div class="header">
      <AlertCircleIcon class="w-5 h-5 text-red-500" />
      Data import failed
    </div>
    <div class="content text-destructive">
      {error}
    </div>
    <div class="footer">
      <Button type="secondary" href={currentFileRoute} onClick={onClose}>
        View YAML
      </Button>
    </div>
  {:else if isDone}
    <div class="header">
      <CheckCircle2Icon class="w-5 h-5 text-green-500" />
      Data imported successfully!
    </div>
    <div class="content">
      <span class="font-mono text-fg-primary break-all">{sourceName}</span>
      has been ingested. What would you like to do next?
    </div>
    <div class="footer">
      <Button type="secondary" href={currentFileRoute} onClick={onClose}>
        View this source
      </Button>

      <Button onClick={generateMetrics} type="primary">
        Generate dashboard

        {#if $ai}
          with AI
          <WandIcon class="w-3 h-3" />
        {/if}
      </Button>
    </div>
  {:else}
    <div class="header">
      <Loader2Icon class="w-5 h-5 text-primary-500 animate-spin" />
      Ingesting data...
    </div>
    <div class="content">
      <p class="font-medium">
        Safe to close this window, we'll notify you when complete.
      </p>
      <p class="mt-2 text-sm text-muted-foreground">
        Processing may take several minutes depending on file size. The upload
        continues in the background, and you'll receive a notification when it's
        complete. You can safely close this modal or cancel at any time.
      </p>
    </div>
    <div class="footer">
      <Button type="secondary" href={currentFileRoute} onClick={onClose}>
        View this source
      </Button>

      <Button onClick={onClose} type="primary">Close</Button>
    </div>
  {/if}
</div>

<style lang="postcss">
  .header {
    @apply flex items-center gap-2;
    @apply text-lg text-fg-primary font-semibold;
  }

  .content {
    @apply text-sm text-fg-muted;
  }

  .footer {
    @apply flex flex-row-reverse gap-2;
  }
</style>
