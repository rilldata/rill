<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

  export let measure: MetricsViewSpecMeasure;
  export let value = "";

  $: name = measure?.displayName || measure?.expression;
</script>

<TooltipContent>
  <div
    class="grid gap-x-4 pointer-events-none pt-1 pb-1 items-baseline"
    style="grid-template-columns: auto max-content"
  >
    <div class="truncate" aria-label="tooltip-name">
      {name}
    </div>

    <div
      class="text-gray-300 justify-self-end"
      aria-label="tooltip-name-description"
    >
      {value}
    </div>
  </div>
</TooltipContent>
