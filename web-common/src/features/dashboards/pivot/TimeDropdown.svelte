<script context="module" lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { getAllowedTimeGrains } from "@rilldata/web-common/lib/time/grains";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { createEventDispatcher } from "svelte";
</script>

<script lang="ts">
  export let selected: string;

  const dispatch = createEventDispatcher();
  const timeControlsStore = useTimeControlStore(getStateManagers());

  $: timeGrainOptions = getAllowedTimeGrains(
    new Date($timeControlsStore.timeStart!),
    new Date($timeControlsStore.timeEnd!),
  );
</script>

<DropdownMenu.Content align="start">
  {#each timeGrainOptions as timeGrain (timeGrain.grain)}
    {@const isSelected = timeGrain.grain === selected}
    <DropdownMenu.Item
      class="flex gap-x-2"
      on:click={() => {
        if (isSelected) return;
        dispatch("select-time-grain", { timeGrain });
      }}
    >
      <span class="w-3 aspect-square">
        {#if isSelected}
          <Check />
        {/if}
      </span>
      {timeGrain.label}
    </DropdownMenu.Item>
  {/each}
</DropdownMenu.Content>
