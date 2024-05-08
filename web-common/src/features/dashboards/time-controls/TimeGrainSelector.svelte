<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import IconSpaceFixer from "@rilldata/web-common/components/button/IconSpaceFixer.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import { isGrainBigger } from "@rilldata/web-common/lib/time/grains";
  import type {
    AvailableTimeGrain,
    TimeGrain,
  } from "@rilldata/web-common/lib/time/types";
  import { createEventDispatcher } from "svelte";
  import type { V1TimeGrain } from "../../../runtime-client";

  export let timeGrainOptions: TimeGrain[];
  export let minTimeGrain: V1TimeGrain;

  const dispatch = createEventDispatcher();
  const timeControlsStore = useTimeControlStore(getStateManagers());

  let open = false;

  $: activeTimeGrain = $timeControlsStore.selectedTimeRange?.interval;
  $: activeTimeGrainLabel =
    activeTimeGrain && TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label;

  $: timeGrains = timeGrainOptions
    .filter((timeGrain) => !isGrainBigger(minTimeGrain, timeGrain.grain))
    .map((timeGrain) => {
      return {
        main: timeGrain.label,
        key: timeGrain.grain,
      };
    });

  const onTimeGrainSelect = (timeGrain: V1TimeGrain) => {
    dispatch("select-time-grain", { timeGrain });
  };
</script>

{#if activeTimeGrain && timeGrainOptions.length}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger asChild let:builder>
      <button
        use:builder.action
        {...builder}
        aria-label="Select a time grain"
        class:bg-gray-200={open}
        class="px-3 py-2 rounded flex flex-row gap-x-2 items-center hover:bg-gray-200 hover:dark:bg-gray-600"
      >
        <div>
          Metric trends by <span class="font-bold">{activeTimeGrainLabel}</span>
        </div>
        <IconSpaceFixer pullRight>
          <div class="transition-transform" class:-rotate-180={open}>
            <CaretDownIcon size="14px" />
          </div>
        </IconSpaceFixer>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-40" align="start">
      {#each timeGrains as option}
        <DropdownMenu.CheckboxItem
          role="menuitem"
          checked={option.key === activeTimeGrain}
          class="text-xs cursor-pointer"
          on:click={() => onTimeGrainSelect(option.key)}
        >
          {option.main}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}
