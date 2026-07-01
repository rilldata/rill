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
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { translateGrainName } from "@rilldata/web-common/lib/time/new-grains";

  export let item: PivotChipData;
  export let removable = false;
  export let grab = false;
  export let slideDuration = 150;
  export let active = false;
  export let fullWidth = false;
  export let onRemove: () => void = () => {};
  export let onmousedown: ((e: MouseEvent) => void) | undefined = undefined;
  export let onclick: ((e: MouseEvent) => void) | undefined = undefined;

  $: activeTimeGrainLabel =
    item.type === PivotChipType.Time && item.id
      ? TIME_GRAIN[item.id as AvailableTimeGrain]?.label
      : undefined;

  // Measure/dimension chips always show a tooltip (display name, plus the
  // description when present). Time chips only have something worth showing
  // when a description is set.
  $: showTooltip = item.type === PivotChipType.Time ? !!item.description : true;
</script>

<Tooltip distance={8} location="top" suppress={!showTooltip} activeDelay={200}>
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
    {onmousedown}
    {onclick}
    {onRemove}
  >
    <div
      slot="body"
      class="flex gap-x-1 items-center justify-start text-left truncate"
    >
      {#if item.type === PivotChipType.Time}
        <b>{m.pivot_time_prefix()}</b>
        {#if activeTimeGrainLabel}
          <p class="grain-label truncate">{translateGrainName(activeTimeGrainLabel)}</p>
        {/if}
      {:else}
        <p class="font-semibold truncate">{item.title}</p>
      {/if}
      <slot name="body" />
    </div>
  </Chip>
  <TooltipContent slot="tooltip-content">
    {#if item.type === PivotChipType.Time}
      {item.description}
    {:else}
      <div class="font-bold">{item.title}</div>
      {#if item.description}
        <div class="text-fg-inverse/70 mt-0.5">{item.description}</div>
      {/if}
    {/if}
  </TooltipContent>
</Tooltip>
