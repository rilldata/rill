<script context="module" lang="ts">
  import { getStateManagers } from "../state-managers/state-managers";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import SearchableFilterDropdown from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterDropdown.svelte";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { potentialFilterName } from "./Filters.svelte";
</script>

<script lang="ts">
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";

  const {
    selectors: {
      dimensions: { allDimensions },
      dimensionFilters: { getAllFilters, isFilterExcludeMode },
    },
  } = getStateManagers();

  function filterExists(name: string, exclude: boolean): boolean {
    const selected = exclude
      ? $getAllFilters?.exclude
      : $getAllFilters?.include;

    return selected?.find((f) => f.name === name) !== undefined;
  }

  $: selectableGroups = [
    <SearchableFilterSelectableGroup>{
      items:
        $allDimensions
          ?.map((d) => ({
            name: d.name as string,
            label: d.label as string,
          }))
          .filter((d) => {
            const exclude = $isFilterExcludeMode(d.name);
            return !filterExists(d.name, exclude);
          }) ?? [],
    },
  ];
</script>

<WithTogglableFloatingElement
  alignment="start"
  distance={8}
  let:active
  let:toggleFloatingElement
>
  <Tooltip distance={8} suppress={active}>
    <button class:active on:click={toggleFloatingElement}>
      <Add size="17px" />
    </button>
    <TooltipContent slot="tooltip-content">Add filter</TooltipContent>
  </Tooltip>

  <SearchableFilterDropdown
    allowMultiSelect={false}
    let:toggleFloatingElement
    on:click-outside={toggleFloatingElement}
    on:escape={toggleFloatingElement}
    on:focus
    on:hover
    on:item-clicked={(e) => {
      toggleFloatingElement();
      $potentialFilterName = e.detail.name;
    }}
    {selectableGroups}
    selectedItems={[]}
    slot="floating-element"
  />
</WithTogglableFloatingElement>

<style lang="postcss">
  button {
    @apply w-[34px] h-[26px] rounded-2xl;
    @apply flex items-center justify-center;
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
