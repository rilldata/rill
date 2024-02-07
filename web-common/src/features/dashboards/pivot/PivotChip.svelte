<script context="module" lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import {
    measureChipColors,
    timeChipColors,
    defaultChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import { createEventDispatcher } from "svelte";
  import type { ChipColors } from "@rilldata/web-common/components/chip/chip-types";
  import type { PivotChipData } from "./types";
  import { PivotChipType } from "./types";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import TimeDropdown from "./TimeDropdown.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";

  const colors: Record<PivotChipType, ChipColors> = {
    time: timeChipColors,
    measure: measureChipColors,
    dimension: defaultChipColors,
  };
</script>

<script lang="ts">
  export let item: PivotChipData;
  export let removable = false;

  const dispatch = createEventDispatcher();

  let open = false;
</script>

{#if item.type === PivotChipType.Time && removable}
  <DropdownMenu.Root bind:open>
    <DropdownMenu.Trigger asChild let:builder>
      <Chip
        outline
        builders={[builder]}
        {removable}
        {...colors[item.type]}
        extraPadding={false}
        extraRounded={true}
        label={item.title}
        on:remove={() => {
          dispatch("remove", item);
        }}
      >
        <div slot="body" class="flex gap-x-1 items-center">
          <b>Time</b>
          <p>{item.title}</p>
          {#if removable}
            <span class="transition-transform" class:-rotate-180={open}>
              <CaretDownIcon size="12px" />
            </span>
          {/if}
        </div>
      </Chip>
    </DropdownMenu.Trigger>
    <TimeDropdown selected={item.id} on:select-time-grain />
  </DropdownMenu.Root>
{:else}
  <Chip
    outline
    {removable}
    {...colors[item.type]}
    extraPadding={false}
    extraRounded={item.type !== PivotChipType.Measure}
    label={item.title}
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
{/if}

<style type="postcss">
  div {
    outline: none !important;
  }
</style>
