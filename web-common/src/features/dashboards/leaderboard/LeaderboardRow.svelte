<script lang="ts">
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LeaderboardValueCell from "./LeaderboardValueCell.svelte";
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import { LeaderboardItemData } from "./leaderboard-utils";
  import { getStateManagers } from "../state-managers/state-managers";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import LeaderboardTooltipContent from "./LeaderboardTooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";

  export let itemData: LeaderboardItemData;
  export let dimensionName: string;
  export let isValidPercentOfTotal: boolean;
  export let isTimeComparisonActive: boolean;
  export let tableWidth: number;
  export let borderTop = false;
  export let borderBottom = false;

  let hovered = false;

  $: selected = itemData.selectedIndex >= 0;

  $: ({
    dimensionValue: label,
    selectedIndex,
    pctOfTotal,
    value: measureValue,
    prevValue: comparisonValue,
  } = itemData);

  const {
    selectors: {
      numberFormat: { activeMeasureFormatter },
      dimensionFilters: { atLeastOneSelection, isFilterExcludeMode },
      comparison: { isBeingCompared: isBeingComparedReadable },
    },
    actions: {
      dimensionsFilter: { toggleDimensionValueSelection },
    },
  } = getStateManagers();

  $: isBeingCompared = $isBeingComparedReadable(dimensionName);
  $: filterExcludeMode = $isFilterExcludeMode(dimensionName);
  $: atLeastOneActive = $atLeastOneSelection(dimensionName);

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: previousValueString =
    comparisonValue !== undefined && comparisonValue !== null
      ? $activeMeasureFormatter(comparisonValue)
      : undefined;

  $: formattedValue = measureValue
    ? $activeMeasureFormatter(measureValue)
    : null;

  $: negativeChange = itemData.deltaAbs !== null && itemData.deltaAbs < 0;

  function shiftClickHandler(label: string) {
    let truncatedLabel = label?.toString();
    if (truncatedLabel?.length > TOOLTIP_STRING_LIMIT) {
      truncatedLabel = `${truncatedLabel.slice(0, TOOLTIP_STRING_LIMIT)}...`;
    }
    copyToClipboard(
      label,
      `copied dimension value "${truncatedLabel}" to clipboard`,
    );
  }
</script>

<tr
  class:border-b={borderBottom}
  class:border-t={borderTop}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  on:click={modified({
    shift: () => shiftClickHandler(label),
    click: (e) =>
      toggleDimensionValueSelection(
        dimensionName,
        label,
        false,
        e.ctrlKey || e.metaKey,
      ),
  })}
>
  <td>
    <LeaderboardItemFilterIcon
      {excluded}
      {isBeingCompared}
      selectionIndex={itemData?.selectedIndex}
    />
  </td>
  <td>
    <Tooltip location="left" distance={20}>
      <LeaderboardValueCell
        {label}
        {comparisonValue}
        {itemData}
        {dimensionName}
        {tableWidth}
      />
      <LeaderboardTooltipContent
        {atLeastOneActive}
        {excluded}
        {filterExcludeMode}
        {label}
        {selected}
        slot="tooltip-content"
      />
    </Tooltip>
  </td>
  <td>
    <div
      class="value-cell flex flex-row items-center gap-x-1 relative whitespace-nowrap z-50"
    >
      {#if previousValueString && hovered}
        <span class="opacity-50">
          {previousValueString} →
        </span>
      {/if}
      <FormattedDataType
        type="INTEGER"
        value={formattedValue || measureValue}
      />
    </div>
  </td>
  {#if isTimeComparisonActive}
    <td>
      <div class="value-cell">
        <FormattedDataType
          type="INTEGER"
          value={itemData.deltaAbs
            ? $activeMeasureFormatter(itemData.deltaAbs)
            : null}
          customStyle={negativeChange ? "text-red-500" : ""}
        />
      </div>
    </td>
    <td>
      <div class="value-cell">
        <PercentageChange
          value={itemData.deltaRel
            ? formatMeasurePercentageDifference(itemData.deltaRel)
            : null}
        />
      </div>
    </td>
  {:else if isValidPercentOfTotal && itemData.pctOfTotal && !isNaN(itemData.pctOfTotal)}
    <td>
      <div class="value-cell">
        <PercentageChange value={itemData.pctOfTotal} />
      </div>
    </td>
  {/if}
</tr>

<style lang="postcss">
  td {
    @apply text-right p-0;
    height: 22px;
  }

  tr {
    @apply cursor-pointer;
  }

  td:first-of-type {
    @apply text-left brightness-100;
  }

  tr:hover {
    @apply brightness-95 bg-background;
  }

  .value-cell {
    @apply pr-2 flex justify-end items-center;
  }
</style>
