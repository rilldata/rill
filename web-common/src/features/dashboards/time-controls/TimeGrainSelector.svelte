<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "../../../runtime-client";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Switch from "@rilldata/web-common/components/button/Switch.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import {
    getAllowedGrains,
    getGrainOrder,
  } from "@rilldata/web-common/lib/time/new-grains";
  const { rillTime } = featureFlags;

  export let tdd = false;
  export let activeTimeGrain: V1TimeGrain | undefined;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;
  export let minTimeGrain: V1TimeGrain | undefined;
  export let usingRillTime: boolean;
  export let complete: boolean;
  export let onTimeGrainSelect: (timeGrain: V1TimeGrain) => void;
  export let toggleComplete: () => void;

  let open = false;
  $: smallestTimeGrainOrder = getGrainOrder(minTimeGrain);
  $: timeGrainOptions = getAllowedGrains(smallestTimeGrainOrder);

  $: activeTimeGrainLabel =
    activeTimeGrain && TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label;

  $: capitalizedLabel = activeTimeGrainLabel
    ?.split(" ")
    .map((word) => {
      return word.charAt(0).toUpperCase() + word.slice(1);
    })
    .join(" ");

  $: timeGrains = minTimeGrain
    ? timeGrainOptions
        .filter((timeGrain) => !isGrainBigger(minTimeGrain, timeGrain.grain))
        .map((timeGrain) => {
          return {
            main: timeGrain.label,
            key: timeGrain.grain,
          };
        })
    : [];
</script>

{#if activeTimeGrain && timeGrainOptions.length && minTimeGrain}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger asChild let:builder>
      <button
        class:tdd
        use:builder.action
        {...builder}
        aria-label="Select a time grain"
        class="flex items-center gap-x-1"
      >
        <div class="items-center flex gap-x-1">
          <span>
            <svelte:element this={tdd ? "b" : "span"}>
              {tdd ? "Time" : "by"}
            </svelte:element>

            <svelte:element this={tdd ? "span" : "b"}>
              {capitalizedLabel}
            </svelte:element>

            {#if complete}
              <i class="ml-0.5">complete</i>
            {/if}
          </span>
          <span class="flex-none transition-transform" class:-rotate-180={open}>
            <CaretDownIcon />
          </span>
        </div>
      </button>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-52" align="start">
      {#each timeGrainOptions as option (option)}
        <DropdownMenu.CheckboxItem
          checkRight
          role="menuitem"
          checked={option === activeTimeGrain}
          class="text-xs cursor-pointer capitalize"
          on:click={() => onTimeGrainSelect(option)}
        >
          {TIME_GRAIN[option].label}
        </DropdownMenu.CheckboxItem>
      {/each}

      {#if $rillTime}
        <DropdownMenu.Separator />
        <div class="flex justify-between px-2 py-1">
          <label for="complete" class="select-none cursor-pointer">
            Complete periods only
          </label>
          <Switch id="complete" checked={complete} on:click={toggleComplete} />
        </div>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<style lang="postcss">
  .tdd {
    @apply border h-7 rounded-full px-2 pl-2.5;
  }

  .tdd:hover {
    @apply bg-gray-50;
  }

  .tdd[data-state="open"] {
    @apply bg-gray-50 border-gray-400;
  }
</style>
