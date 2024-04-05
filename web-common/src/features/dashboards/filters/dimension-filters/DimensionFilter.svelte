<script lang="ts">
  import {
    defaultChipColors,
    excludeChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import { getDimensionType } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/getDimensionType";
  import { STRING_LIKES } from "@rilldata/web-common/lib/duckdb-data-types";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
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
    metricsViewName,
  } = StateManagers;

  $: isInclude = !$dashboardStore.dimensionFilterExcludeMode.get(name);

  let isOpen = false;
  let searchText = "";
  let allValues: Record<string, string[]> = {};
  let topListQuery: ReturnType<typeof getFilterSearchList> | undefined;

  $: dimensionType = getDimensionType(
    $runtime.instanceId,
    $metricsViewName,
    name,
  );
  $: stringLikeDimension = STRING_LIKES.has($dimensionType.data ?? "");

  $: if (isOpen) {
    topListQuery = getFilterSearchList(StateManagers, {
      dimension: name,
      searchText,
      addNull: searchText.length !== 0 && "null".includes(searchText),
      type: $dimensionType.data,
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
  allValues={allValues[name]}
  colors={getColorForChip(isInclude)}
  enableSearch={stringLikeDimension}
  excludeMode={!isInclude}
  label="View filter"
  name={isInclude ? label : `Exclude ${label}`}
  on:apply
  on:click={() => setOpen()}
  on:mount={() => setOpen()}
  on:remove
  on:search={(event) => {
    handleSearch(event.detail);
  }}
  on:toggle={() => toggleDimensionFilterMode(name)}
  {selectedValues}
  typeLabel="dimension"
>
  <svelte:fragment slot="body-tooltip-content">
    Click to edit the the filters in this dimension
  </svelte:fragment>
</RemovableListChip>
