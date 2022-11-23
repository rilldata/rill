<script lang="ts">
  import LeaderboardListItem from "$lib/components/leaderboard/LeaderboardListItem.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import TooltipTitle from "$lib/components/tooltip/TooltipTitle.svelte";
  import {
    formatBigNumberPercentage,
    formatInteger,
  } from "$lib/util/formatters";
  import {
    useRuntimeServiceGetTableCardinality,
    useRuntimeServiceGetTopK,
  } from "@rilldata/web-common/runtime-client";
  import { LIST_SLIDE_DURATION } from "@rilldata/web-local/lib/application-config";
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    copyToClipboard,
    createShiftClickAction,
  } from "@rilldata/web-local/lib/util/shift-click-action";
  import { format } from "d3-format";
  import { slide } from "svelte/transition";
  import Shortcut from "../../../tooltip/Shortcut.svelte";
  import StackingWord from "../../../tooltip/StackingWord.svelte";
  import TooltipShortcutContainer from "../../../tooltip/TooltipShortcutContainer.svelte";
  export let objectName: string;
  export let columnName: string;

  const { shiftClickAction } = createShiftClickAction();

  let topK;
  let sliceAmount = 15;

  $: topKQuery = useRuntimeServiceGetTopK(
    $runtimeStore?.instanceId,
    objectName,
    columnName,
    {
      agg: "count(*)",
      k: 75,
    }
  );
  $: topK = $topKQuery?.data?.categoricalSummary?.topK?.entries;

  /**
   * Get the total rows for this profile.
   */
  let totalRowsQuery;
  $: totalRowsQuery = useRuntimeServiceGetTableCardinality(
    $runtimeStore?.instanceId,
    objectName
  );
  // FIXME: count should not be a string.
  $: totalRowsString = $totalRowsQuery?.data?.cardinality;
  $: totalRows = +totalRowsString;

  $: smallestPercentage =
    topK && topK.length
      ? Math.min(...topK?.slice(0, 5)?.map((entry) => entry.count / totalRows))
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
</script>

{#if topK && totalRows}
  <div transition:slide|local={{ duration: LIST_SLIDE_DURATION }} class="py-4">
    {#each topK.slice(0, sliceAmount) as item (item.value)}
      {@const negligiblePercentage = item.count / totalRows < 0.0002}
      {@const percentage = negligiblePercentage
        ? "<.01%"
        : formatPercentage(item.count / totalRows)}
      <LeaderboardListItem compact value={item.count / totalRows}>
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
                <svelte:fragment slot="name">{columnName}</svelte:fragment>
                <svelte:fragment slot="description"
                  >{formatBigNumberPercentage(item.count / totalRows)} of rows</svelte:fragment
                >
              </TooltipTitle>
              <TooltipShortcutContainer>
                <div>
                  <StackingWord key="shift">copy</StackingWord> column value to clipboard
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
                <svelte:fragment slot="name">{columnName}</svelte:fragment>
                <svelte:fragment slot="description"
                  >{formatBigNumberPercentage(item.count / totalRows)} of rows</svelte:fragment
                >
              </TooltipTitle>
              <TooltipShortcutContainer>
                <div>
                  <StackingWord key="shift">copy</StackingWord> count to clipboard
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
