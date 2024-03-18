<script lang="ts">
  import {
    chartPromptsStore,
    ChartPromptStatus,
  } from "@rilldata/web-common/features/charts/prompt/chartPrompt";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let chartName: string;

  $: chartPrompt = chartPromptsStore.getStatusForChart(chartName);
</script>

{#if $chartPrompt && $chartPrompt.status !== ChartPromptStatus.Idle}
  <div class="flex flex-row gap-x-2">
    <Spinner size="16px" status={EntityStatus.Running} />
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
