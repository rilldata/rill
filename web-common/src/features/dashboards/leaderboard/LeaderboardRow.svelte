<script lang="ts">
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LeaderboardValueCell from "./LeaderboardValueCell.svelte";
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import { LeaderboardItemData } from "./leaderboard-utils";
  import { getStateManagers } from "../state-managers/state-managers";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";

  export let itemData: LeaderboardItemData;
  export let dimensionName: string;
  export let isPercentOfTotal: boolean;
  export let isTimeComparisonActive: boolean;

  $: label = itemData.dimensionValue;
  $: measureValue = itemData.value;
  $: selected = itemData.selectedIndex >= 0;
  $: comparisonValue = itemData.prevValue;
  $: pctOfTotal = itemData.pctOfTotal;

  const {
    selectors: {
      numberFormat: { activeMeasureFormatter },
      activeMeasure: { isSummableMeasure },
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

  $: formattedValue = measureValue
    ? $activeMeasureFormatter(measureValue)
    : null;

  $: previousValueString =
    comparisonValue !== undefined && comparisonValue !== null
      ? $activeMeasureFormatter(comparisonValue)
      : undefined;
  $: showPreviousTimeValue = hovered && previousValueString !== undefined;
  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: renderedBarValue = $isSummableMeasure && pctOfTotal ? pctOfTotal : 0;

  $: color = excluded
    ? "ui-measure-bar-excluded"
    : selected
      ? "ui-measure-bar-included-selected"
      : "ui-measure-bar-included";

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

  let hovered = false;
  const onHover = () => {
    hovered = true;
  };
  const onLeave = () => {
    hovered = false;
  };
</script>

<tr>
  <td>
    <LeaderboardItemFilterIcon
      {excluded}
      {isBeingCompared}
      selectionIndex={itemData?.selectedIndex}
    />
  </td>
  <td>
    <LeaderboardValueCell {itemData} {dimensionName} />
  </td>
  <td>
    <FormattedDataType type="INTEGER" value={itemData.value} />
  </td>
  {#if isTimeComparisonActive}
    <td>
      <div class="value-cell">
        <FormattedDataType type="INTEGER" value={itemData.deltaRel} />
      </div></td
    >
    <td
      ><div class="value-cell">
        <FormattedDataType type="INTEGER" value={itemData.deltaAbs} />
      </div>
    </td>
  {:else if isPercentOfTotal}
    <td><div class="value-cell">{itemData.pctOfTotal}</div></td>
  {/if}
</tr>

<style lang="postcss">
  td {
    @apply text-right truncate;
    height: 22px;
  }

  td:first-of-type {
    @apply text-left;
  }

  tr:hover {
    @apply bg-gray-100;
  }

  td {
    @apply p-0;
  }

  .value-cell {
    @apply px-1;
  }
</style>
