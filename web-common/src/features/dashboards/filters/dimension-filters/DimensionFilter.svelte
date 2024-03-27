<script lang="ts">
  import {
    defaultChipColors,
    excludeChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import RemovableListChip from "../../../../components/chip/removable-list-chip/RemovableListChip.svelte";
  import { getFilterSearchList } from "../../selectors/index";
  import { getStateManagers } from "../../state-managers/state-managers";

  export let name: string;
  export let label: string;
  export let column: string;
  export let selectedValues: string[];

  const StateManagers = getStateManagers();
  const {
    dashboardStore,
    actions: {
      dimensionsFilter: { toggleDimensionFilterMode },
    },
  } = StateManagers;

  $: isInclude = !$dashboardStore.dimensionFilterExcludeMode.get(name);

  let isOpen = false;
  let searchText = "";
  let allValues: Record<string, string[]> = {};
  let topListQuery: ReturnType<typeof getFilterSearchList> | undefined;

  $: if (isOpen) {
    topListQuery = getFilterSearchList(StateManagers, {
      dimension: name,
      searchText,
      addNull: "null".includes(searchText),
    });
  }

  $: if (!$topListQuery?.isFetching) {
    const topListData = $topListQuery?.data?.data ?? [];
    allValues[name] = topListData.map((datum) => datum[column]) ?? [];
  }

  function getColorForChip(isInclude: boolean) {
    return isInclude ? defaultChipColors : excludeChipColors;
  }

  function setOpen() {
    isOpen = true;
  }

  function handleSearch(value: string) {
    searchText = value;
  }
</script>

<RemovableListChip
  on:remove
  on:apply
  on:mount={() => setOpen()}
  on:click={() => setOpen()}
  on:toggle={() => toggleDimensionFilterMode(name)}
  on:search={(event) => {
    handleSearch(event.detail);
  }}
  typeLabel="dimension"
  name={isInclude ? label : `Exclude ${label}`}
  excludeMode={!isInclude}
  colors={getColorForChip(isInclude)}
  label="View filter"
  {selectedValues}
  allValues={allValues[name]}
>
  <svelte:fragment slot="body-tooltip-content">
    Click to edit the the filters in this dimension
  </svelte:fragment>
</RemovableListChip>
