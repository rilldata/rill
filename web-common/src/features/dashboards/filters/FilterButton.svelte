<script context="module" lang="ts">
  import { getStateManagers } from "../state-managers/state-managers";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import SearchableFilterDropdown from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterDropdown.svelte";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
</script>

<script lang="ts">
  const {
    selectors: {
      dimensions: { allDimensions },
    },
    actions: {
      dimensionsFilter: { toggleDimensionValueSelection },
    },
  } = getStateManagers();

  $: selectableItems = $allDimensions
    ? $allDimensions.map((d) => ({
        name: d.name as string,
        label: d.label as string,
      }))
    : [];
</script>

<WithTogglableFloatingElement
  distance={8}
  alignment="start"
  let:toggleFloatingElement
  let:active
>
  <Tooltip distance={8} suppress={active}>
    <button class:active on:click={toggleFloatingElement}>
      <Add size="17px" />
    </button>
    <TooltipContent slot="tooltip-content">Add filter</TooltipContent>
  </Tooltip>

  <SearchableFilterDropdown
    slot="floating-element"
    selectedItems={[]}
    allowMultiSelect={false}
    {selectableItems}
    on:escape={toggleFloatingElement}
    on:click-outside={toggleFloatingElement}
    on:item-clicked={(e) => {
      toggleFloatingElement();
      toggleDimensionValueSelection(e.detail.name);
    }}
  />
</WithTogglableFloatingElement>

<style lang="postcss">
  button {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply px-[8px] py-[4px];
    @apply border border-dashed border-slate-300;
  }

  button:hover {
    @apply bg-slate-100;
  }

  button:active,
  .active {
    @apply bg-slate-200;
  }
</style>
