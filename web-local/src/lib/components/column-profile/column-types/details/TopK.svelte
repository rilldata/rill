<script lang="ts">
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import StackingWord from "@rilldata/web-common/components/tooltip/StackingWord.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import TooltipTitle from "@rilldata/web-common/components/tooltip/TooltipTitle.svelte";
  import LeaderboardListItem from "@rilldata/web-common/features/dashboards/leaderboard/LeaderboardListItem.svelte";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-common/layout/config";
  import {
    copyToClipboard,
    createShiftClickAction,
  } from "@rilldata/web-common/lib/actions/shift-click-action";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "@rilldata/web-common/lib/formatters";
  import type { TopKEntry } from "@rilldata/web-common/runtime-client";
  import { format } from "d3-format";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";

  export let colorClass = "bg-blue-200";

  const { shiftClickAction } = createShiftClickAction();

  export let topK: TopKEntry[];
  export let totalRows: number;
  export let k = 15;

  const dispatch = createEventDispatcher();

  $: smallestPercentage =
    topK && topK.length
      ? Math.min(...topK.slice(0, 5).map((entry) => entry.count / totalRows))
      : undefined;
  $: formatPercentage =
    smallestPercentage < 0.01
      ? format("0.2%")
      : smallestPercentage
      ? format("0.1%")
      : () => "";

  function ensureSpaces(str: string, n = 6) {
    return `${Array.from({ length: n - str.length })
      .fill("&nbsp;")
      .join("")}${str}`;
  }

  let tooltipProps = { location: "right", distance: 16 };

  function handleFocus(value: TopKEntry) {
    return () => dispatch("focus-top-k", value);
  }

  function handleBlur(value: TopKEntry) {
    return () => dispatch("blur-top-k", value);
  }

  /** handle LISTs and STRUCTs */
</script>

{#if topK && totalRows}
  <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }}>
    {#each topK.slice(0, k) as item (item.value)}
      {@const negligiblePercentage = item.count / totalRows < 0.0002}
      {@const percentage = negligiblePercentage
        ? "<.01%"
        : formatPercentage(item.count / totalRows)}
      <LeaderboardListItem
        compact
        value={item.count / totalRows}
        color={colorClass}
        showIcon={false}
        on:focus={handleFocus(item)}
        on:blur={handleBlur(item)}
      >
        <svelte:fragment slot="title">
          <Tooltip {...tooltipProps}>
            <div
              style:font-size="12px"
              class="text-ellipsis overflow-hidden whitespace-nowrap"
              use:shiftClickAction
              on:shift-click={() =>
                copyToClipboard(
                  item.value,
                  `copied column value "${
                    item.value === null ? "NULL" : item.value
                  }" to clipboard`
                )}
            >
              {item.value}
            </div>
            <TooltipContent slot="tooltip-content">
              <TooltipTitle>
                <svelte:fragment slot="name"
                  >{`${item.value}`.slice(0, 100)}</svelte:fragment
                >
                <svelte:fragment slot="description"
                  >{formatBigNumberPercentage(item.count / totalRows)} of rows</svelte:fragment
                >
              </TooltipTitle>
              <TooltipShortcutContainer>
                <div>
                  <StackingWord key="shift">Copy</StackingWord> column value to clipboard
                </div>
                <Shortcut>
                  <span style="font-family: var(--system);">⇧</span> + Click
                </Shortcut>
              </TooltipShortcutContainer>
            </TooltipContent>
          </Tooltip>
        </svelte:fragment>
        <svelte:fragment slot="right">
          <Tooltip {...tooltipProps}>
            <div
              use:shiftClickAction
              on:shift-click={() =>
                copyToClipboard(
                  item.count,
                  `copied ${item.count} to clipboard`
                )}
            >
              {formatInteger(item.count)}
              <span class="ui-copy-inactive pl-2">
                {@html ensureSpaces(percentage)}</span
              >
            </div>
            <TooltipContent slot="tooltip-content">
              <TooltipTitle>
                <svelte:fragment slot="name"
                  >{`${item.value}`.slice(0, 100)}</svelte:fragment
                >
                <svelte:fragment slot="description"
                  >{formatBigNumberPercentage(item.count / totalRows)} of rows</svelte:fragment
                >
              </TooltipTitle>
              <TooltipShortcutContainer>
                <div>
                  <StackingWord key="shift">Copy</StackingWord> count to clipboard
                </div>
                <Shortcut>
                  <span style="font-family: var(--system);">⇧</span> + Click
                </Shortcut>
              </TooltipShortcutContainer>
            </TooltipContent>
          </Tooltip>
        </svelte:fragment>
      </LeaderboardListItem>
    {/each}
  </div>
{/if}
