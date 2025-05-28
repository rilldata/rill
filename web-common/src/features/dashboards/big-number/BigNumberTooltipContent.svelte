<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipDescription from "@rilldata/web-common/components/tooltip/TooltipDescription.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";

  export let measure: MetricsViewSpecMeasure;
  export let value = "";

  $: description =
    measure?.description || measure?.displayName || measure?.expression;
  $: name = measure?.displayName || measure?.expression;
</script>

<TooltipContent maxWidth="280px">
  <div
    class="grid gap-x-2 pointer-events-none pt-1 pb-1 items-baseline"
    style="grid-template-columns: auto max-content"
    style:min-width="200px"
  >
    <div class="font-normal truncate" aria-label="tooltip-name">
      {name}
    </div>
    <div
      class="text-gray-300 justify-self-end pl-3"
      style:max-width="280px"
      aria-label="tooltip-name-description"
    >
      {value}
    </div>
  </div>

  <TooltipDescription>
    {description}
  </TooltipDescription>
</TooltipContent>
