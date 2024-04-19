<script lang="ts">
  import CancelCircle from "@rilldata/web-common/components/icons/CancelCircle.svelte";
  import {
    chartPromptsStore,
    ChartPromptStatus,
  } from "@rilldata/web-common/features/charts/prompt/chartPrompt";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let chartName: string;

  $: chartPrompt = chartPromptsStore.getStatusForChart(chartName);
</script>

<!-- TODO: handle prompt error -->
<div class="size-full flex items-center justify-center">
  {#if $chartPrompt && $chartPrompt?.status === ChartPromptStatus.Error}
    <div class="flex flex-col gap-y-2 text-red-600 items-center justify-center">
      <CancelCircle size="16px" />
      <div class="flex flex-col gap-y-2">
        <div>Failed to generate chart using AI</div>
        <div>Using prompt: "{$chartPrompt.prompt}"</div>
        <div>{$chartPrompt.error}</div>
      </div>
    </div>
  {:else if $chartPrompt && $chartPrompt.status !== ChartPromptStatus.Idle}
    <div class="flex flex-col gap-y-2 items-center justify-center">
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
</div>
