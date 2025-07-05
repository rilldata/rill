<script context="module" lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
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
  export let withDropdown = false;
  export let availableGrains: V1TimeGrain[] = [];
  export let onRemove: () => void = () => {};
  export let onTimeGrainSelect: (timeGrain: V1TimeGrain) => void = () => {};

  let dropdownOpen = false;

  $: timeGrainOptions =
    withDropdown && item.type === PivotChipType.Time
      ? availableGrains.map((grain) => ({
          main: TIME_GRAIN[grain as AvailableTimeGrain]?.label || grain,
          key: grain,
        }))
      : [];

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

  function handleTimeGrainSelect(timeGrain: V1TimeGrain) {
    onTimeGrainSelect(timeGrain);
    dropdownOpen = false;
  }
</script>

{#if withDropdown && item.type === PivotChipType.Time && timeGrainOptions.length > 0}
  <DropdownMenu.Root bind:open={dropdownOpen}>
    <DropdownMenu.Trigger asChild let:builder>
      <div use:builder.action {...builder}>
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
              <b>Time</b>
              <div class="grain-dropdown flex items-center gap-x-1">
                <p>{capitalizedLabel || item.title}</p>
                <span
                  class="flex-none transition-transform"
                  class:-rotate-180={dropdownOpen}
                >
                  <CaretDownIcon size="12px" />
                </span>
              </div>
            </div>
          </Chip>
          <TooltipContent slot="tooltip-content">
            {item.description}
          </TooltipContent>
        </Tooltip>
      </div>
    </DropdownMenu.Trigger>
    <DropdownMenu.Content class="min-w-52" align="start">
      {#each timeGrainOptions as option (option.key)}
        <DropdownMenu.CheckboxItem
          checkRight
          role="menuitem"
          checked={option.key === item.id}
          class="text-xs cursor-pointer capitalize"
          on:click={() => handleTimeGrainSelect(option.key)}
        >
          {option.main}
        </DropdownMenu.CheckboxItem>
      {/each}
    </DropdownMenu.Content>
  </DropdownMenu.Root>
{:else}
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
          <p>{capitalizedLabel || item.title}</p>
        {:else}
          <p class="font-semibold">{item.title}</p>
        {/if}
      </div>
    </Chip>
    <TooltipContent slot="tooltip-content">
      {item.description}
    </TooltipContent>
  </Tooltip>
{/if}
