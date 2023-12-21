<script context="module" lang="ts">
  import { getStateManagers } from "../state-managers/state-managers";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Add from "@rilldata/web-common/components/icons/Add.svelte";
  import SearchableFilterDropdown from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterDropdown.svelte";
  import WithTogglableFloatingElement from "@rilldata/web-common/components/floating-element/WithTogglableFloatingElement.svelte";
  import type { SearchableFilterSelectableItem } from "@rilldata/web-common/components/searchable-filter-menu/SearchableFilterSelectableItem";
  import { getHavingFilterExpression } from "@rilldata/web-common/features/dashboards/state-managers/selectors/measure-filters";
  import { createBinaryExpression } from "@rilldata/web-common/features/dashboards/stores/filter-generators";
  import { V1Operation } from "@rilldata/web-common/runtime-client";
  import { potentialFilterName } from "./Filters.svelte";
</script>

<script lang="ts">
  const {
    dashboardStore,
    selectors: {
      measures: { allMeasures },
      dimensions: { allDimensions },
      dimensionFilters: { dimensionHasFilter },
    },
    actions: {
      measuresFilter: { toggleMeasureFilter },
    },
  } = getStateManagers();

  let isDimension: Record<string, boolean> = {};
  let selectableItems: Array<SearchableFilterSelectableItem> = [];
  if ($allDimensions && $allMeasures) {
    isDimension = {};
    selectableItems = [];

    $allDimensions
      .map((d) => ({
        name: d.name as string,
        label: d.label as string,
      }))
      .filter((d) => !$dimensionHasFilter(d.name))
      .forEach((d) => {
        selectableItems.push(d);
        isDimension[d.name] = true;
      });

    $allMeasures
      .map((m) => ({
        name: m.name as string,
        label: m.label as string,
      }))
      .filter((m) => {
        return (
          getHavingFilterExpression({ dashboard: $dashboardStore })(m.name) ===
          undefined
        );
      })
      .forEach((m) => {
        selectableItems.push(m);
        isDimension[m.name] = false;
      });
  }

  function filterAdded(name: string) {
    const dimIdx = $allDimensions?.findIndex((d) => d.name === name);
    if (dimIdx === undefined || dimIdx === -1) {
      toggleMeasureFilter(
        name,
        createBinaryExpression(name, V1Operation.OPERATION_LT, 0)
      );
    } else {
      toggleDimensionNameSelection(name);
    }
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
    {selectableItems}
    selectedItems={[]}
    slot="floating-element"
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
