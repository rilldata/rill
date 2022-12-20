<script lang="ts">
  import LeaderboardListItem from "$lib/components/leaderboard/LeaderboardListItem.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "$lib/util/formatters";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import {
    copyToClipboard,
    createShiftClickAction,
  } from "@rilldata/web-local/lib/util/shift-click-action";
  import { format } from "d3-format";
  import { createEventDispatcher } from "svelte";
  import { slide } from "svelte/transition";
  import Shortcut from "../../../tooltip/Shortcut.svelte";
  import StackingWord from "../../../tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "../../../tooltip/TooltipShortcutContainer.svelte";
  export let colorClass = "bg-blue-200";

  const { shiftClickAction } = createShiftClickAction();

  export let topK;
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

  function ensureSpaces(str, n = 6) {
    return `${Array.from({ length: n - str.length })
      .fill("&nbsp;")
      .join("")}${str}`;
  }

  let tooltipProps = { location: "right", distance: 16 };

  function handleFocus(value) {
    return () => dispatch("focus-top-k", value);
  }

  function handleBlur(value) {
    return () => dispatch("blur-top-k", value);
  }
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
