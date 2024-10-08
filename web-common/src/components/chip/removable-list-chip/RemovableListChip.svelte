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

  import { Chip } from "../index";
  import RemovableListBody from "./RemovableListBody.svelte";
  import RemovableListMenu from "./RemovableListMenu.svelte";
</script>

<script lang="ts">
  export let name: string;
  export let selectedValues: string[];
  export let allValues: string[] | null;
  export let enableSearch = true;
  export let type: "measure" | "dimension" | "time";
  /** an optional type label that will appear in the tooltip */
  export let typeLabel: string;
  export let excludeMode: boolean;

  export let label: string | undefined = undefined;

  let active = !selectedValues.length;

  const dispatch = createEventDispatcher();

  onMount(() => {
    dispatch("mount");
  });

  function handleDismiss() {
    if (!selectedValues.length) {
      dispatch("remove");
    } else {
      active = false;
    }
  }
</script>

<WithTogglableFloatingElement
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
      {type}
      {active}
      {label}
      exclude={excludeMode}
      on:click={() => {
        toggleFloatingElement();
        dispatch("click");
      }}
      caret
      on:remove={() => dispatch("remove")}
      removable
      removeTooltipText={`remove ${selectedValues.length} value${
        selectedValues.length !== 1 ? "s" : ""
      } for ${name}`}
    >
      <RemovableListBody
        slot="body"
        label={name}
        show={1}
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

        <slot name="body-tooltip-content">click to edit the values</slot>
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
</WithTogglableFloatingElement>
