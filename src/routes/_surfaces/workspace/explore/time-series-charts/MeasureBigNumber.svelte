<script lang="ts">
  import { WithTween } from "$lib/components/data-graphic/functional-components";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  export let value: number;
  export let formatter: (value: number) => string = undefined;
  export let description: string = undefined;
</script>

<div>
  <Tooltip location="top" distance={16}>
    <h2>
      <slot name="name" />
    </h2>
    <TooltipContent slot="tooltip-content">
      {description}
    </TooltipContent>
  </Tooltip>
  <div style:font-size="1.5rem" style:font-weight="light" class="text-gray-600">
    <!-- the default slot will be a tweened number that uses the formatter. One can optionally
    override this by filling the slot in the consuming component. -->
    <slot name="value">
      <WithTween {value} tweenProps={{ duration: 500 }} let:output>
        {formatter ? formatter(output) : output}
      </WithTween>
    </slot>
  </div>
</div>
