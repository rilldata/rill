<script lang="ts">
  import FormattedDataType from "@rilldata/web-common/components/data-types/FormattedDataType.svelte";
  import PercentageChange from "@rilldata/web-common/components/data-types/PercentageChange.svelte";
  import ExternalLink from "@rilldata/web-common/components/icons/ExternalLink.svelte";
  import { TOOLTIP_STRING_LIMIT } from "@rilldata/web-common/layout/config";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { modified } from "@rilldata/web-common/lib/actions/modified-click";
  import { clamp } from "@rilldata/web-common/lib/clamp";
  import { formatMeasurePercentageDifference } from "@rilldata/web-common/lib/number-formatting/percentage-formatter";
  import { slide } from "svelte/transition";
  import { type LeaderboardItemData, makeHref } from "./leaderboard-utils";
  import { cellInspectorStore } from "../stores/cell-inspector-store";
  import LeaderboardItemFilterIcon from "./LeaderboardItemFilterIcon.svelte";
  import LongBarZigZag from "./LongBarZigZag.svelte";
  import {
    COMPARISON_COLUMN_WIDTH,
    DEFAULT_COLUMN_WIDTH,
    valueColumn,
    deltaColumn,
    MEASURES_PADDING,
  } from "./leaderboard-widths";

  export let itemData: LeaderboardItemData;
  export let dimensionName: string;
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
    ? "var(--color-gray-100)"
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
  style:background={leaderboardMeasureNames.length === 1
    ? dimensionGradients
    : undefined}
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
  <td
    role="button"
    tabindex="0"
    data-dimension-cell
    class:ui-copy={!atLeastOneActive}
    class:ui-copy-disabled={excluded}
    class:ui-copy-strong={!excluded && selected}
    on:click={modified({
      shift: () => shiftClickHandler(dimensionValue),
    })}
    on:pointerover={() => {
      if (dimensionValue) {
        // Always update the value in the store, but don't change visibility
        cellInspectorStore.updateValue(dimensionValue.toString());
      }
    }}
    on:focus={() => {
      if (dimensionValue) {
        // Always update the value in the store, but don't change visibility
        cellInspectorStore.updateValue(dimensionValue.toString());
      }
    }}
    class="relative size-full flex flex-none justify-between items-center leaderboard-label"
    style:background={dimensionGradients}
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
  </td>

  {#each Object.keys(values) as measureName}
    <td
      role="button"
      tabindex="0"
      data-measure-cell
      on:click={modified({
        shift: () => shiftClickHandler(values[measureName]?.toString() || ""),
      })}
      style:background={leaderboardMeasureNames.length === 1
        ? measureGradients
        : measureGradientMap?.[measureName]}
      on:pointerover={() => {
        const value = values[measureName]?.toString() || "";
        if (value) {
          cellInspectorStore.updateValue(value);
        }
      }}
      on:focus={() => {
        const value = values[measureName]?.toString() || "";
        if (value) {
          cellInspectorStore.updateValue(value);
        }
      }}
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
    </td>

    {#if isValidPercentOfTotal(measureName) && shouldShowContextColumns(measureName)}
      <td
        role="button"
        tabindex="0"
        data-comparison-cell
        on:click={modified({
          shift: () =>
            shiftClickHandler(pctOfTotals[measureName]?.toString() || ""),
        })}
        on:pointerover={() => {
          const value = pctOfTotals[measureName]?.toString() || "";
          if (value) {
            cellInspectorStore.updateValue(value);
          }
        }}
        on:focus={() => {
          const value = pctOfTotals[measureName]?.toString() || "";
          if (value) {
            cellInspectorStore.updateValue(value);
          }
        }}
      >
        <PercentageChange
          value={pctOfTotals[measureName]}
          color="text-gray-500"
        />
        {#if showZigZags[measureName]}
          <LongBarZigZag />
        {/if}
      </td>
    {/if}

    {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
      <td
        role="button"
        tabindex="0"
        data-comparison-cell
        on:click={modified({
          shift: () =>
            shiftClickHandler(deltaAbsMap[measureName]?.toString() || ""),
        })}
        on:pointerover={() => {
          const value = deltaAbsMap[measureName]?.toString() || "";
          if (value) {
            cellInspectorStore.updateValue(value);
          }
        }}
        on:focus={() => {
          const value = deltaAbsMap[measureName]?.toString() || "";
          if (value) {
            cellInspectorStore.updateValue(value);
          }
        }}
      >
        <FormattedDataType
          color="text-gray-500"
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
      </td>
    {/if}

    {#if isTimeComparisonActive && shouldShowContextColumns(measureName)}
      <td
        data-comparison-cell
        on:click={modified({
          shift: () =>
            shiftClickHandler(deltaRels[measureName]?.toString() || ""),
        })}
        on:pointerover={() => {
          const value = deltaRels[measureName]?.toString() || "";
          if (value) {
            cellInspectorStore.updateValue(value);
          }
        }}
        on:focus={() => {
          const value = deltaRels[measureName]?.toString() || "";
          if (value) {
            cellInspectorStore.updateValue(value);
          }
        }}
      >
        <PercentageChange
          value={deltaRels[measureName]
            ? formatMeasurePercentageDifference(deltaRels[measureName])
            : null}
          color="text-gray-500"
        />
        {#if showZigZags[measureName]}
          <LongBarZigZag />
        {/if}
      </td>
    {/if}
  {/each}
</tr>

<style lang="postcss">
  td {
    @apply text-right p-0;
    @apply px-2 relative;
    height: 22px;
  }

  tr {
    @apply cursor-pointer;
    max-height: 22px;
  }

  tr:hover {
    @apply bg-gray-100;
  }

  td[data-comparison-cell] {
    @apply bg-surface px-1 truncate;
  }

  td[data-dimension-cell] {
    @apply sticky left-0 z-30 bg-surface;
  }

  tr:hover td[data-dimension-cell],
  tr:hover td[data-comparison-cell] {
    @apply bg-gray-100;
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

  td {
    height: 22px !important;
  }
</style>
