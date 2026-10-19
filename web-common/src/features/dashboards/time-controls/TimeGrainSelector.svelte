<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import { translateGrainName } from "@rilldata/web-common/lib/time/new-grains";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "../../../runtime-client";

  export let tdd = false;
  /** The grain in effect. In adaptive mode this is the grain derived from the time range. */
  export let activeTimeGrain: V1TimeGrain | undefined;
  export let timeStart: string | undefined;
  export let timeEnd: string | undefined;
  export let minTimeGrain: V1TimeGrain | undefined;
  /** Called with `undefined` when the user picks adaptive mode. */
  export let onTimeGrainSelect: (timeGrain: V1TimeGrain | undefined) => void;
  /** Whether the grain currently follows the time range. */
  export let adaptiveTimeGrain = false;
  /** Shows the adaptive option in the menu. Off for contexts that need a fixed grain. */
  export let allowAdaptiveTimeGrain = false;
  export let complete: boolean = false;
  export let side: "top" | "right" | "bottom" | "left" = "bottom";

  let open = false;

  $: timeGrainOptions =
    timeStart && timeEnd
      ? getAllowedTimeGrains(new Date(timeStart), new Date(timeEnd))
      : [];
  $: activeTimeGrainLabel =
    activeTimeGrain && TIME_GRAIN[activeTimeGrain as AvailableTimeGrain]?.label;

  $: capitalizedLabel = activeTimeGrainLabel
    ? translateGrainName(activeTimeGrainLabel)
        .split(" ")
        .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
        .join(" ")
    : undefined;

  $: timeGrains = minTimeGrain
    ? timeGrainOptions
        .filter((timeGrain) => !isGrainBigger(minTimeGrain, timeGrain.grain))
        .map((timeGrain) => {
          return {
            main: translateGrainName(timeGrain.label),
            key: timeGrain.grain,
          };
        })
    : [];
</script>

{#if activeTimeGrain && timeGrainOptions.length && minTimeGrain}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger>
      {#snippet child({ props })}
        <button
          {...props}
          class:tdd
          aria-label={m.dashboard_select_time_grain()}
          class="flex items-center gap-x-1"
        >
          <div class="items-center flex gap-x-1">
            <span>
              <svelte:element this={tdd ? "b" : "span"}>
                {tdd ? m.time_grain_time() : m.time_grain_by()}
              </svelte:element>

              <svelte:element this={tdd ? "span" : "b"}>
                {capitalizedLabel}
              </svelte:element>

              {#if complete}
                <i class="ml-0.5">{m.time_grain_complete()}</i>
              {/if}
              {#if allowAdaptiveTimeGrain && adaptiveTimeGrain}
                <i class="ml-0.5 text-fg-muted"
                  >{m.time_grain_adaptive_hint()}</i
                >
              {/if}
            </span>
            <span
              class="flex-none transition-transform"
              class:-rotate-180={open}
            >
              <CaretDownIcon />
            </span>
          </div>
        </button>
      {/snippet}
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-52" align="start" {side}>
      {#if allowAdaptiveTimeGrain}
        <DropdownMenu.CheckboxItem
          checkRight
          closeOnSelect
          checked={adaptiveTimeGrain}
          class="text-xs cursor-pointer"
          onSelect={() => {
            onTimeGrainSelect(undefined);
            open = false;
          }}
        >
          <div class="flex flex-col">
            <span>{m.time_grain_adaptive()}</span>
            <span class="text-fg-muted">
              {m.time_grain_adaptive_description()}
            </span>
          </div>
        </DropdownMenu.CheckboxItem>
        <DropdownMenu.Separator />
      {/if}
      {#each timeGrains as option (option.key)}
        <DropdownMenu.CheckboxItem
          checkRight
          closeOnSelect
          checked={!adaptiveTimeGrain && option.key === activeTimeGrain}
          class="text-xs cursor-pointer capitalize"
          onSelect={() => {
            onTimeGrainSelect(option.key);
            open = false;
          }}
        >
          {option.main}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{/if}

<style lang="postcss">
  .tdd {
    @apply border h-7 rounded-full px-2 pl-2.5 bg-surface-background text-fg-primary;
  }

  .tdd:hover {
    @apply bg-surface-background;
  }

  .tdd:global([data-state="open"]) {
    @apply bg-surface-background border-gray-400;
  }
</style>
