<!-- @component 

Implementation for showing a Chip with a title + a list of values. Shows the first value in selectedValues, while enabling
a way to select values from a menu popover.

The RemovableListChip has a few features that are worth noting:
- the remove toggle is on the left side, rather than the right side, which is more traditional
with chips. The main reason for this is the user should not have to look to the left side of a longer chip to see
the name and then move the cursor to the right to cancel it.
- clicking the chip body will expand out a the RemovableListMenu. This component will be in charge of both selecting / de-selecting
existing elements in the lib as well as changing the type (include, exclude) and enabling list search. The implementation of these parts
are details left to the consumer of the component; this component should remain pure-ish (only internal state) if possible.
-->

<script context="module" lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { fly } from "svelte/transition";
  import WithTogglableFloatingElement from "../../floating-element/WithTogglableFloatingElement.svelte";
  import Tooltip from "../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../tooltip/TooltipContent.svelte";
  import TooltipTitle from "../../tooltip/TooltipTitle.svelte";
  import { ChipColors, defaultChipColors } from "../chip-types";
  import { Chip } from "../index";
  import RemovableListBody from "./RemovableListBody.svelte";
  import RemovableListMenu from "./RemovableListMenu.svelte";
  import * as DropdownMenu from "../../dropdown-menu/";
  // import { createEventDispatcher } from "svelte";
  import Cancel from "../../icons/Cancel.svelte";
  import Check from "../../icons/Check.svelte";
  import Spacer from "../../icons/Spacer.svelte";
  import { Menu, MenuItem } from "../../menu";
  import { Search } from "../../search";
  import Footer from "./Footer.svelte";
  import Button from "../../button/Button.svelte";
</script>

<script lang="ts">
  export let name: string;
  export let selectedValues: string[];
  export let allValues: string[];
  export let enableSearch = true;
  export let openOnMount = false;

  /** an optional type label that will appear in the tooltip */
  export let typeLabel: string;
  export let excludeMode: boolean;
  export let colors: ChipColors = defaultChipColors;
  export let label: string | undefined = undefined;

  let active = false;

  const dispatch = createEventDispatcher();

  onMount(() => {
    dispatch("mount");
    active = openOnMount;
  });

  function handleDismiss() {
    if (!selectedValues.length) {
      dispatch("remove");
    } else {
      active = false;
    }
  }

  function onSearch() {
    dispatch("search", searchText);
  }

  function toggleValue(value: string) {
    dispatch("apply", value);
  }

  let allSelected = false;
  let searchText = "";
</script>

<DropdownMenu.Root
  typeahead={false}
  bind:open={active}
  closeOnItemClick={false}
  onOpenChange={(open) => {
    if (open) {
      searchText = "";
    }
  }}
>
  <DropdownMenu.Trigger asChild let:builder>
    <Tooltip
      activeDelay={60}
      alignment="start"
      distance={8}
      location="bottom"
      suppress={active}
    >
      <Chip
        builders={[builder]}
        {...colors}
        {active}
        {label}
        on:remove={() => dispatch("remove")}
        outline
        removable
      >
        <svelte:fragment slot="remove-tooltip">
          <slot name="remove-tooltip-content">
            remove {selectedValues.length}
            value{#if selectedValues.length !== 1}s{/if} for {name}</slot
          >
        </svelte:fragment>

        <RemovableListBody
          {active}
          label={name}
          show={1}
          slot="body"
          values={selectedValues}
        />
      </Chip>
      <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
        <TooltipContent maxWidth="400px">
          <TooltipTitle>
            <svelte:fragment slot="name">{name}</svelte:fragment>
            <svelte:fragment slot="description"
              >{typeLabel || ""}</svelte:fragment
            >
          </TooltipTitle>
          {#if $$slots["body-tooltip-content"]}
            <slot name="body-tooltip-content">click to edit the values</slot>
          {/if}
        </TooltipContent>
      </div>
    </Tooltip>
  </DropdownMenu.Trigger>

  <DropdownMenu.Content
    align="start"
    class="flex flex-col max-h-96 w-72 overflow-hidden p-0"
  >
    {#if enableSearch}
      <div class="px-3 py-2 pt-3">
        <Search
          bind:value={searchText}
          on:input={onSearch}
          label="Search list"
          showBorderOnFocus={false}
        />
      </div>
    {/if}

    <div class="flex flex-col flex-1 overflow-y-auto w-full h-fit pb-1">
      {#each allValues.sort() as value (value)}
        <DropdownMenu.Item
          class="flex gap-x-2"
          on:click={() => {
            toggleValue(value);
          }}
        >
          {#if selectedValues.includes(value) && !excludeMode}
            <Check size="20px" color="#15141A" />
          {:else if selectedValues.includes(value) && excludeMode}
            <Cancel size="20px" color="#15141A" />
          {:else}
            <Spacer size="20px" />
          {/if}

          <span
            class:ui-copy-disabled={selectedValues.includes(value) &&
              excludeMode}
          >
            {#if value?.length > 240}
              {value.slice(0, 240)}...
            {:else}
              {value}
            {/if}
          </span>
        </DropdownMenu.Item>
      {:else}
        <div
          class="ui-copy-disabled text-center justify-center h-8 items-center flex"
        >
          no results
        </div>
      {/each}
    </div>
    <Footer>
      <Button on:click={() => {}} type="text">
        {#if allSelected}
          Deselect all
        {:else}
          Select all
        {/if}
      </Button>

      <Button on:click={() => dispatch("toggle")} type="secondary">
        {#if excludeMode}
          Include
        {:else}
          Exclude
        {/if}
      </Button>
    </Footer>
  </DropdownMenu.Content>
</DropdownMenu.Root>

<!-- <WithTogglableFloatingElement
  alignment="start"
  bind:active
  distance={8}
  let:toggleFloatingElement
>
  <Tooltip
    activeDelay={60}
    alignment="start"
    distance={8}
    location="bottom"
    suppress={active}
  >
    <Chip
      {...colors}
      {active}
      {label}
      on:click={() => {
        toggleFloatingElement();
        dispatch("click");
      }}
      on:remove={() => dispatch("remove")}
      outline
      removable
    >

      <svelte:fragment slot="remove-tooltip">
        <slot name="remove-tooltip-content">
          remove {selectedValues.length}
          value{#if selectedValues.length !== 1}s{/if} for {name}</slot
        >
      </svelte:fragment>

      <RemovableListBody
        {active}
        label={name}
        show={1}
        slot="body"
        values={selectedValues}
      />
    </Chip>
    <div slot="tooltip-content" transition:fly={{ duration: 100, y: 4 }}>
      <TooltipContent maxWidth="400px">
        <TooltipTitle>
          <svelte:fragment slot="name">{name}</svelte:fragment>
          <svelte:fragment slot="description">{typeLabel || ""}</svelte:fragment
          >
        </TooltipTitle>
        {#if $$slots["body-tooltip-content"]}
          <slot name="body-tooltip-content">click to edit the values</slot>
        {/if}
      </TooltipContent>
    </div>
  </Tooltip>
  <RemovableListMenu
    {allValues}
    {enableSearch}
    {excludeMode}
    on:apply
    on:click-outside={handleDismiss}
    on:escape={handleDismiss}
    on:search
    on:toggle
    {selectedValues}
    slot="floating-element"
  />
</WithTogglableFloatingElement> -->
