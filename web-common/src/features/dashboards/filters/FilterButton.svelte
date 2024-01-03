<script context="module" lang="ts">
  import { getStateManagers } from "../state-managers/state-managers";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import SearchableFilterDropdown from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterDropdown.svelte";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import { potentialFilterName } from "./filter-items";
</script>

<script lang="ts">
  import type { SearchableFilterSelectableGroup } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import {
    getDimensionDisplayName,
    getMeasureDisplayName,
  } from "@rilldata/web-common/features/dashboards/filters/getDisplayName";

  const {
    selectors: {
      measures: { allMeasures },
      measureFilters: { measureHasFilter },
      dimensions: { allDimensions },
      dimensionFilters: { dimensionHasFilter },
    },
  } = getStateManagers();

  let isDimension: Record<string, boolean> = {};
  let selectableGroups: SearchableFilterSelectableGroup[] = [];
  $: if ($allDimensions && $allMeasures) {
    isDimension = {};
    selectableGroups = [];

    const measureSelectableGroup = <SearchableFilterSelectableGroup>{
      name: "MEASURES",
      items: [],
    };
    $allMeasures
      .map((m) => ({
        name: m.name as string,
        label: getMeasureDisplayName(m),
      }))
      .filter((m) => !$measureHasFilter(m.name))
      .forEach((m) => {
        measureSelectableGroup.items.push(m);
        isDimension[m.name] = false;
      });
    selectableGroups.push(measureSelectableGroup);

    const dimensionSelectableGroup = <SearchableFilterSelectableGroup>{
      name: "DIMENSIONS",
      items: [],
    };
    $allDimensions
      .map((d) => ({
        name: d.name as string,
        label: getDimensionDisplayName(d),
      }))
      .filter((d) => !$dimensionHasFilter(d.name))
      .forEach((d) => {
        dimensionSelectableGroup.items.push(d);
        isDimension[d.name] = true;
      });
    selectableGroups.push(dimensionSelectableGroup);
  }
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
