<script lang="ts">
  import { measureSelection } from "@rilldata/web-common/features/dashboards/time-series/measure-selection/measure-selection.ts";
  import { Bot } from "lucide-svelte";

  export let measureName: string;
  export let metricsViewName: string;

  $: ({ measure, x, y } = measureSelection);

  $: forThisMeasure = $measure === measureName;
  $: correctedX = ($x ?? 0) - 35;
  $: correctedY = -($y ?? 0);

  function onExplain(e) {
    e.stopPropagation();
    e.preventDefault();
    measureSelection.startAnomalyExplanationChat(metricsViewName);
  }
</script>

{#if forThisMeasure}
  <div class="relative">
    <div
      class="flex flex-row gap-x-1 items-center absolute"
      style="left: {correctedX}px;top: {correctedY}px;"
      on:click={onExplain}
      role="presentation"
      tabindex="-1"
    >
      <Bot size={16} class="stroke-primary cursor-pointer -mt-0.5" />
      <span
        role="presentation"
        class="fill-primary stroke-surface cursor-pointer hover:underline"
      >
        Explain (E)
      </span>
    </div>
  </div>
{/if}
