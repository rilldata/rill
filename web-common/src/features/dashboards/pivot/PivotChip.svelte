<script context="module" lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { createEventDispatcher } from "svelte";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
</script>

<script lang="ts">
  export let item: PivotChipData;
  export let removable = false;
  export let grab = false;
  export let slideDuration = 150;
  export let active = false;

  const dispatch = createEventDispatcher();
</script>

<Chip
  type={item.type}
  label={item.title}
  caret={false}
  {grab}
  {active}
  {slideDuration}
  {removable}
  supressTooltip
  on:mousedown
  on:click
  on:remove={() => {
    dispatch("remove", item);
  }}
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
