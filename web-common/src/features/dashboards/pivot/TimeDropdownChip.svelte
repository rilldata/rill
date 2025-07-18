<script context="module" lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
  import type { AvailableTimeGrain } from "@rilldata/web-common/lib/time/types";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import PivotChip from "./PivotChip.svelte";
  import type { PivotChipData } from "./types";
</script>

<script lang="ts">
  export let item: PivotChipData;
  export let removable = false;
  export let grab = false;
  export let slideDuration = 150;
  export let active = false;
  export let fullWidth = false;
  export let availableGrains: V1TimeGrain[] = [];
  export let onTimeGrainSelect: (timeGrain: V1TimeGrain) => void = () => {};
  export let onRemove: () => void = () => {};

  let dropdownOpen = false;

  $: timeGrainOptions = availableGrains.map((grain) => ({
    main: TIME_GRAIN[grain as AvailableTimeGrain]?.label || grain,
    key: grain,
  }));

  function handleTimeGrainSelect(timeGrain: V1TimeGrain) {
    onTimeGrainSelect(timeGrain);
    dropdownOpen = false;
  }
</script>

{#if timeGrainOptions.length > 0}
  <DropdownMenu.Root bind:open={dropdownOpen}>
    <DropdownMenu.Trigger asChild let:builder>
      <div use:builder.action {...builder}>
        <PivotChip
          {item}
          {removable}
          {grab}
          {active}
          {slideDuration}
          {fullWidth}
          on:mousedown
          on:click
          {onRemove}
        >
          <div class="grain-dropdown" slot="body">
            <span
              class="flex-none transition-transform"
              class:-rotate-180={dropdownOpen}
            >
              <CaretDownIcon size="12px" />
            </span>
          </div>
        </PivotChip>
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
  <PivotChip
    {item}
    {removable}
    {grab}
    {active}
    {slideDuration}
    {fullWidth}
    on:mousedown
    on:click
    {onRemove}
  />
{/if}
