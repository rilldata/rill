<script lang="ts">
  import {
    type AddDataConfig,
    type ImportAddDataStep,
    ImportDataStep,
  } from "@rilldata/web-common/features/add-data/manager/steps/types.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
  import {
    getFileHref,
    withEditorPrefix,
  } from "@rilldata/web-common/layout/navigation/editor-routing";
  import { previewModeStore } from "@rilldata/web-common/layout/preview-mode-store";
  import { runImportSteps } from "@rilldata/web-common/features/add-data/manager/steps/import.ts";
  import type { AddDataStateManager } from "@rilldata/web-common/features/add-data/manager/AddDataStateManager.svelte.ts";

  export let config: AddDataConfig;
  export let stateManager: AddDataStateManager;
  export let importAddDataStep: ImportAddDataStep;
  export let onDone: () => void;

  const { ai, developerChat } = featureFlags;

  const runtimeClient = useRuntimeClient();

  let importStep = ImportDataStep.Init;
  $: currentFileRoute = $previewModeStore
    ? withEditorPrefix("/dashboards")
    : withEditorPrefix("/");
  $: sourceName = importAddDataStep.config.importTo.modelName ?? "";
  $: isDone = importStep === ImportDataStep.Done;
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
      await runImportSteps(
        runtimeClient,
        config,
        importAddDataStep,
        (step, currentFilePath) => {
          importStep = step;
          if (currentFilePath) {
            if ($previewModeStore) {
              currentFileRoute = withEditorPrefix("/dashboards");
            } else {
              currentFileRoute = getFileHref(addLeadingSlash(currentFilePath));
            }
          }
        },
      );
    } catch (e) {
      error = e?.response?.data?.message ?? e?.message ?? "Unknown error";
      stateManager.fireErrorEvent(error!, importStep);
    }
  }

  async function generateMetrics() {
    onDone();
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
      <Button type="secondary" href={currentFileRoute} onClick={onDone}>
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
      <Button onClick={generateMetrics} type="primary">
        Generate dashboard

        {#if $ai}
          with AI
          <WandIcon class="w-3 h-3" />
        {/if}
      </Button>

      <Button type="secondary" href={currentFileRoute} onClick={onDone}>
        View this source
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
        complete. You can safely close this window — the import will continue in
        the background.
      </p>
    </div>
    <div class="footer">
      <Button type="secondary" href={currentFileRoute} onClick={onDone}>
        View this source
      </Button>

      <Button onClick={onDone} type="primary">Close</Button>
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
