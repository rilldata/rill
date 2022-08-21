<!-- @component 
The main feature implementation for an individual FilterSet Chip component.default

The FilterSet has a few features that are worth noting:
- the cancel toggle is on the left side, rather than the right side, which is more traditional
with chips. The main reason for this is the user should not have to look to the left side of a longer chip to see
the name and then move the cursor to the right to cancel it.
- clicking the chip body will expand out a the FilterMenu. This component will be in charge of both selecting / de-selecting
existing filters as well as changing the type (include, exclude) and enabling dimension search.
-->
<script>
  import { Chip } from "$lib/components/chip";
  import WithTogglableFloatingElement from "$lib/components/floating-element/WithTogglableFloatingElement.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";
  import FilterMenu from "./FilterMenu.svelte";
  import FilterSetBody from "./FilterSetBody.svelte";

  export let name;
  export let selectedValues;

  const dispatch = createEventDispatcher();

  let active = false;
</script>

<WithTogglableFloatingElement
  let:toggleFloatingElement
  bind:active
  distance={8}
  alignment="start"
>
  <Tooltip
    location="bottom"
    alignment="start"
    distance={8}
    activeDelay={60}
    suppress={active}
  >
    <Chip
      removable
      on:click={toggleFloatingElement}
      on:remove={() => dispatch("remove-filters")}
      {active}
    >
      <!-- remove button tooltip -->
      <svelte:fragment slot="remove-tooltip"
        >remove {selectedValues.length}
        {name}
        dimension filter{#if selectedValues.length !== 1}s{/if}</svelte:fragment
      >
      <!-- body -->
      <FilterSetBody
        slot="body"
        label={name}
        values={selectedValues}
        show={1}
      />
    </Chip>
    <div slot="tooltip-content" transition:fly|local={{ duration: 100, y: 4 }}>
      <TooltipContent maxWidth="400px">
        <TooltipTitle>
          <svelte:fragment slot="name">{name}</svelte:fragment>
          <svelte:fragment slot="description">dimension</svelte:fragment>
        </TooltipTitle>
        click to edit the filters in this dimension
      </TooltipContent>
    </div>
  </Tooltip>
  <FilterMenu
    slot="floating-element"
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    on:close={toggleFloatingElement}
    on:select
    {selectedValues}
  />
</WithTogglableFloatingElement>
