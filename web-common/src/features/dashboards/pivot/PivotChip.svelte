<script context="module" lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
</script>

<script lang="ts">
  export let item: PivotChipData;
  export let removable = false;
  export let grab = false;
  export let slideDuration = 150;
  export let active = false;
  export let fullWidth = false;
  export let onRemove: () => void = () => {};

  $: activeTimeGrainLabel =
    item.type === PivotChipType.Time && item.id
      ? TIME_GRAIN[item.id as AvailableTimeGrain]?.label
      : undefined;

  $: capitalizedLabel = activeTimeGrainLabel
    ?.split(" ")
    .map((word) => {
      return word.charAt(0).toUpperCase() + word.slice(1);
    })
    .join(" ");
</script>

<Tooltip
  distance={8}
  location="right"
  suppress={!item.description}
  activeDelay={200}
>
  <Chip
    theme
    type={item.type}
    label="{item.title} pivot chip"
    caret={false}
    {grab}
    {active}
    {slideDuration}
    {removable}
    {fullWidth}
    supressTooltip
    on:mousedown
    on:click
    {onRemove}
  >
    <div slot="body" class="flex gap-x-1 items-center">
      {#if item.type === PivotChipType.Time}
        <b>Time</b>
        {#if capitalizedLabel}
          <p class="grain-label">{capitalizedLabel}</p>
        {/if}
      {:else}
        <p class="font-semibold">{item.title}</p>
      {/if}
      <slot name="body" />
    </div>
  </Chip>
  <TooltipContent slot="tooltip-content">
    {item.description}
  </TooltipContent>
</Tooltip>
