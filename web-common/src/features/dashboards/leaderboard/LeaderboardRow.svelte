<script lang="ts">
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { slide } from "svelte/transition";
  import { type LeaderboardItemData, makeHref } from "./leaderboard-utils";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import {
    COMPARISON_COLUMN_WIDTH,
    DEFAULT_COLUMN_WIDTH,
    valueColumn,
    deltaColumn,
    MEASURES_PADDING,
  } from "./leaderboard-widths";
  import LeaderboardCell from "@rilldata/web-common/features/dashboards/leaderboard/LeaderboardCell.svelte";

  export let itemData: LeaderboardItemData;
  export let dimensionName: string;
  export let dataType: string;
  export let borderTop = false;
  export let borderBottom = false;
  export let isBeingCompared: boolean;
  export let filterExcludeMode: boolean;
  export let atLeastOneActive: boolean;
  export let isTimeComparisonActive: boolean;
  export let leaderboardMeasureNames: string[] = [];
  export let leaderboardShowContextForAllMeasures: boolean;
  export let leaderboardSortByMeasureName: string | null;
  export let isValidPercentOfTotal: (measureName: string) => boolean;
  export let dimensionColumnWidth: number;
  export let maxValues: Record<string, number> = {};
  export let toggleDimensionValueSelection: (
    dimensionName: string,
    dimensionValue: string,
    keepPillVisible?: boolean | undefined,
    isExclusiveFilter?: boolean | undefined,
  ) => void;
  export let formatters: Record<
    string,
    (value: number | string | null | undefined) => string | null | undefined
  >;

  function shouldShowContextColumns(measureName: string): boolean {
    return (
      leaderboardShowContextForAllMeasures ||
      measureName === leaderboardSortByMeasureName
    );
  }

  let hovered = false;
  let valueRect = new DOMRect(0, 0, DEFAULT_COLUMN_WIDTH);
  let deltaRect = new DOMRect(0, 0, COMPARISON_COLUMN_WIDTH);
  let parent: HTMLTableRowElement;

  $: ({
    dimensionValue,
    selectedIndex,
    values,
    pctOfTotals,
    prevValues,
    deltaRels,
    deltaAbs: deltaAbsMap,
    uri,
  } = itemData);

  $: selected = selectedIndex >= 0;

  // Super important special case: if there is not at least one "active" (selected) value,
  // we need to set *all* items to be included, because by default if a user has not
  // selected any values, we assume they want all values included in all calculations.
  $: excluded = atLeastOneActive
    ? (filterExcludeMode && selected) || (!filterExcludeMode && !selected)
    : false;

  $: previousValueString =
    leaderboardMeasureNames.length === 1 &&
    prevValues[leaderboardMeasureNames[0]] !== undefined &&
    prevValues[leaderboardMeasureNames[0]] !== null
      ? formatters[leaderboardMeasureNames[0]]?.(
          prevValues[leaderboardMeasureNames[0]] as number,
        )
      : undefined;

  $: href = makeHref(uri, dimensionValue);

  $: deltaElementWidth = deltaRect?.width;
  $: valueElementWith = valueRect?.width;

  $: valueColumn.update(valueElementWith);
  $: deltaColumn.update(deltaElementWidth);

  $: barColor = excluded
    ? "var(--surface-active)"
    : selected || hovered
      ? "var(--color-theme-200)"
      : "var(--color-theme-100)";

  // Calculate bar width excluding dimension column
  $: barWidth =
    $valueColumn +
    (leaderboardMeasureNames.length === 1
      ? dimensionColumnWidth
      : -MEASURES_PADDING);
  // Calculate bar lengths based on max value. For percent-of-total measure, this will be the total.
  $: barLengths = Object.fromEntries(
    Object.entries(values).map(([name, value]) => {
      const maxValue = maxValues[name];
      if (!value || !maxValue || maxValue <= 0) return [name, 0];
      // Calculate relative magnitude: current value / max value * available bar width
      return [name, (Math.abs(value) / maxValue) * barWidth];
    }),
  );

  $: totalBarLength = Object.values(barLengths).reduce(
    (sum, length) => sum + length,
    0,
  );

  $: showZigZags = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      length > barWidth,
    ]),
  );

  $: measureCellBarLengths = Object.fromEntries(
    Object.entries(barLengths).map(([name, length]) => [
      name,
      clamp(0, length, $valueColumn),
    ]),
  );

  $: dimensionGradients =
    leaderboardMeasureNames.length === 1
      ? `linear-gradient(to right, ${barColor} ${Math.min(dimensionColumnWidth, totalBarLength)}px, transparent ${Math.min(dimensionColumnWidth, totalBarLength)}px)`
      : undefined;

  $: measureGradients =
    leaderboardMeasureNames.length === 1
      ? `linear-gradient(to right, ${barColor} ${Math.max(0, totalBarLength - dimensionColumnWidth)}px, transparent ${Math.max(0, totalBarLength - dimensionColumnWidth)}px)`
      : undefined;

  $: measureGradientMap =
    leaderboardMeasureNames.length === 1
      ? undefined
      : Object.fromEntries(
          leaderboardMeasureNames.map((name) => {
            const length = measureCellBarLengths[name];
            return [
              name,
              length
                ? `linear-gradient(to right, transparent ${MEASURES_PADDING}px, ${barColor} ${MEASURES_PADDING}px, ${barColor} ${length + MEASURES_PADDING}px, transparent ${length + MEASURES_PADDING}px)`
                : undefined,
            ];
          }),
        );

  $: dimensionCellClass = `relative size-full flex flex-none justify-between items-center leaderboard-label ${
    atLeastOneActive ? "cursor-pointer" : ""
  } ${excluded ? "text-fg-disabled" : ""} ${!excluded && selected ? "text-fg-primary font-semibold" : ""}`;

  function onDimensionCellClick(e: MouseEvent) {
    // Check if user has selected text
    const selection = window.getSelection();
    if (selection && selection.toString().length > 0) {
      // User has selected text, don't trigger row selection
      return;
    }

    // If no text is selected, proceed with normal click behavior
    toggleDimensionValueSelection(
      dimensionName,
      dimensionValue,
      false,
      e.ctrlKey || e.metaKey,
    );
  }
</script>

<tr
  bind:this={parent}
  class:border-b={borderBottom}
  class:border-t={borderTop}
  class="relative"
  on:pointerover={() => (hovered = true)}
  on:pointerout={() => (hovered = false)}
  on:click={(e) => {
    if (e.shiftKey) return;
    onDimensionCellClick(e);
  }}
>
  <td data-comparison-cell>
    <LeaderboardItemFilterIcon
      {excluded}
      {isBeingCompared}
      selectionIndex={itemData?.selectedIndex}
    />
  </td>
  <LeaderboardCell
    value={dimensionValue}
    {dataType}
    cellType="dimension"
    className={dimensionCellClass}
    background={dimensionGradients}
  >
    <span class="truncate select-text">
      <FormattedDataType value={dimensionValue} truncate />
    </span>

    {#if previousValueString && hovered}
      <span
        class="opacity-50 whitespace-nowrap font-normal"
        transition:slide={{ axis: "x", duration: 200 }}
      >
        {previousValueString} →
      </span>
    {/if}

    {#if href}
      <span class="external-link-wrapper">
        <a
          target="_blank"
          rel="noopener noreferrer"
          {href}
          title={href}
          on:click|stopPropagation
          class:hovered
        >
          <ExternalLink className="fill-primary-600" />
        </a>
      </span>
    {/if}
  </LeaderboardCell>

  {#each Object.keys(values) as measureName, i (i)}
    <LeaderboardCell
      value={values[measureName]?.toString() || ""}
      dataType="INTEGER"
      cellType="measure"
      background={leaderboardMeasureNames.length === 1
        ? measureGradients
        : measureGradientMap?.[measureName]}
    >
      <div class="w-fit ml-auto bg-transparent" bind:contentRect={valueRect}>
        <FormattedDataType
          type="INTEGER"
          value={values[measureName]
            ? formatters[measureName]?.(values[measureName])
            : null}
        />
      </div>

      {#if showZigZags[measureName] && !isTimeComparisonActive && !isValidPercentOfTotal(measureName)}
        <LongBarZigZag />
      {/if}
    </LeaderboardCell>

    {#if isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)}
      <LeaderboardCell
        value={pctOfTotals[measureName]?.toString() || ""}
        dataType="INTEGER"
        cellType="comparison"
      >
        <PercentageChange
          value={pctOfTotals[measureName]}
          color="text-fg-secondary"
        />
        {#if showZigZags[measureName]}
          <LongBarZigZag />
        {/if}
      </LeaderboardCell>
    {/if}

    {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
      <LeaderboardCell
        value={deltaAbsMap[measureName]?.toString() || ""}
        dataType="INTEGER"
        cellType="comparison"
      >
        <FormattedDataType
          color="text-fg-secondary"
          type="INTEGER"
          value={deltaAbsMap[measureName]
            ? formatters[measureName]?.(deltaAbsMap[measureName])
            : null}
          customStyle={deltaAbsMap[measureName] !== null &&
          deltaAbsMap[measureName] < 0
            ? "text-red-500"
            : ""}
          truncate={true}
        />
      </LeaderboardCell>
    {/if}

    {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
      <LeaderboardCell
        value={deltaRels[measureName]?.toString() || ""}
        {dataType}
        cellType="comparison"
      >
        <PercentageChange
          value={deltaRels[measureName]
            ? formatMeasurePercentageDifference(deltaRels[measureName])
            : null}
          color="text-fg-secondary"
        />
        {#if showZigZags[measureName]}
          <LongBarZigZag />
        {/if}
      </LeaderboardCell>
    {/if}
  {/each}
</tr>

<style lang="postcss">
  td {
    @apply h-[22px] p-0 px-1 truncate text-right;
  }

  tr {
    @apply cursor-pointer;
    max-height: 22px;
  }

  tr:hover {
    @apply bg-popover-accent;
  }

  td[data-comparison-cell] {
    @apply bg-transparent px-1 truncate;
  }

  tr:hover td[data-comparison-cell] {
    @apply bg-popover-accent;
  }

  .external-link-wrapper a {
    opacity: 0;
    position: absolute;
    right: 0;
    top: 0;
    height: 20px;
    width: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .external-link-wrapper a.hovered {
    opacity: 0.7;
    pointer-events: auto;
    backdrop-filter: blur(2px);
    -webkit-backdrop-filter: blur(2px);
  }
</style>
