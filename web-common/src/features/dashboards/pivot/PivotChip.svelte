<script context="module" lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
</script>

<script lang="ts">
  export let item: PivotChipData;
  export let removable = false;
  export let grab = false;
  export let slideDuration = 150;
  export let active = false;
  export let fullWidth = false;
  export let onRemove: () => void = () => {};
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
    on:remove={onRemove}
  >
    <div slot="body" class="flex gap-x-1 items-center">
      {#if item.type === PivotChipType.Time}
        <b>Time</b>
        <p>{item.title}</p>
      {:else}
        <p class="font-semibold">{item.title}</p>
      {/if}
    </div>
  </Chip>
  <TooltipContent slot="tooltip-content">
    {item.description}
  </TooltipContent>
</Tooltip>
