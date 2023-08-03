<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import {
    ChipColors,
    excludeChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";

  export let start: Date;
  export let end: Date;
  export let zone: string;

  const dispatch = createEventDispatcher();
  const colors: ChipColors = excludeChipColors;
  const label = "Subrange";

  $: orderedStart = getOrderedStartEnd(start, end)?.start;
  $: orderedEnd = getOrderedStartEnd(start, end)?.end;
</script>

<Chip
  removable
  on:click
  on:remove={() => dispatch("remove")}
  active={false}
  {...colors}
  {label}
>
  Hello
  <!-- remove button tooltip -->
  <svelte:fragment slot="remove-tooltip">
    <slot name="remove-tooltip-content">remove selected subrange</slot>
  </svelte:fragment>

  <div slot="body" class="flex gap-x-2">
    <div class="font-bold text-ellipsis overflow-hidden whitespace-nowrap">
      {label}
    </div>
    <div class="flex flex-wrap gap-x-2 gap-y-1">
      {prettyFormatTimeRange(
        orderedStart,
        orderedEnd,
        TimeRangePreset.CUSTOM,
        zone
      )}
    </div>
  </div>
</Chip>
