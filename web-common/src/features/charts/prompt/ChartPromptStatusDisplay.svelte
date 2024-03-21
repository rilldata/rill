<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import {
    chartPromptsStore,
    ChartPromptStatus,
  } from "@rilldata/web-common/features/charts/prompt/chartPrompt";

  export let chartName: string;

  $: chartPrompt = chartPromptsStore.getStatusForChart(chartName);
</script>

<!-- TODO: handle prompt error -->
{#if $chartPrompt && $chartPrompt?.status === ChartPromptStatus.Error}
  <div class="flex flex-row gap-x-2 text-red-600">
    <CancelCircle size="16px" />
    <div class="flex flex-col gap-y-2">
      <div>Failed to generate chart using AI</div>
      <div>Using prompt: "{$chartPrompt.prompt}"</div>
      <div>{$chartPrompt.error}</div>
    </div>
  </div>
{:else if $chartPrompt && $chartPrompt.status !== ChartPromptStatus.Idle}
  <div class="flex flex-row gap-x-2">
    <CancelCircle size="16px" />
    <div class="flex flex-col gap-y-2">
      <div>
        Generating {$chartPrompt.status === ChartPromptStatus.GeneratingData
          ? "data"
          : "chart spec"} using AI
      </div>
      <div>Using prompt: "{$chartPrompt.prompt}"</div>
    </div>
  </div>
{:else}
  <slot />
{/if}
