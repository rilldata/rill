<script lang="ts">
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { fly } from "svelte/transition";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { MetricsViewSpecMeasureV2 } from "@rilldata/web-common/runtime-client";

  export let tooltipText: string;
  export let measures: MetricsViewSpecMeasureV2[];
  export let firstMeasure: MetricsViewSpecMeasureV2;
  export let selectedMeasureCount: number = 1;
  export let onToggle: (count: number) => void;

  let inputValue = selectedMeasureCount.toString();

  function handleInputChange(event: Event) {
    const input = event.target as HTMLInputElement;
    const value = parseInt(input.value);
    if (value > 0 && value <= measures.length) {
      onToggle(value);
    }
  }
</script>

<Tooltip activeDelay={60} alignment="start" distance={8} location="bottom">
  <Button type="text" label={firstMeasure.displayName || firstMeasure.name}>
    <div
      class="flex items-center gap-x-1 px-1 text-gray-700 hover:text-inherit font-normal"
    >
      Showing
      <input
        type="number"
        min="1"
        max={measures.length}
        bind:value={inputValue}
        on:change={handleInputChange}
        class="w-12 px-2 py-1 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      <strong>leaderboard measures</strong>
    </div>
  </Button>

  <div slot="tooltip-content" transition:fly={{ duration: 300, y: 4 }}>
    <TooltipContent maxWidth="400px">
      {tooltipText}
    </TooltipContent>
  </div>
</Tooltip>
