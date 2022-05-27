<script lang="ts">
  /**
   * LeaderboardFeature.svelte
   * -------------------------
   * This is the "implemented" feature of the leaderboard, meant to be used
   * in the application itself.
   */
  import { createEventDispatcher } from "svelte";
  import { fly } from "svelte/transition";

  import LeaderboardContainer from "$lib/components/leaderboard/LeaderboardContainer.svelte";
  import LeaderboardHeader from "$lib/components/leaderboard/LeaderboardHeader.svelte";
  import LeaderboardList from "$lib/components/leaderboard/LeaderboardList.svelte";
  import LeaderboardListItem from "$lib/components/leaderboard/LeaderboardListItem.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import Expand from "$lib/components/icons/ExpandCaret.svelte";

  import { formatBigNumberPercentage } from "$lib/util/formatters";

  export let displayName;
  /** The reference value is the one that the bar in the LeaderboardListItem
   * gets scaled with. For a summable metric, the total is a reference value,
   * or for a count(*) metric, the reference value is the total number of rows.
   */
  export let referenceValue: number;
  export let values;
  export let activeValues: string[];

  export let slice = 7;
  export let seeMoreSlice = 50;

  export let seeMore = false;

  const dispatch = createEventDispatcher();

  $: atLeastOneActive = !!activeValues?.length;

  let righthandElements = [];
  $: widths = righthandElements.map(
    (element) => element?.getBoundingClientRect()?.width || 0
  );

  /** figure out how many selected values are currently hidden */
  // $: hiddenSelectedValues = values.filter((di, i) => {
  //   return activeValues.includes(di.label) && i > slice - 1 && !seeMore;
  // });

  let expanded = false;
</script>

<LeaderboardContainer focused={atLeastOneActive}>
  <Tooltip location="top" alignment="start" distance={4}>
    <LeaderboardHeader isActive={atLeastOneActive}>
      <div
        slot="title"
        class:text-gray-500={atLeastOneActive}
        class:italic={atLeastOneActive}
      >
        {displayName}
      </div>
      <div slot="right">
        <button
          on:click={() => {
            dispatch("expand");
          }}
          >{#if expanded}less{:else}<Expand size={16} />{/if}</button
        >
      </div>
    </LeaderboardHeader>
    <TooltipContent slot="tooltip-content">
      {#if activeValues.length}
        filtering {displayName} by {activeValues.length} value{#if activeValues.length !== 1}s{/if}
      {:else}
        click on the fields to filter by ____
      {/if}
    </TooltipContent>
  </Tooltip>
  <LeaderboardList>
    {#each values.slice(0, !seeMore ? slice : seeMoreSlice) as { label, value }, i (label)}
      {@const isActive = activeValues?.includes(label)}
      <div>
        <Tooltip location="right">
          <LeaderboardListItem
            value={referenceValue ? value / referenceValue : 0}
            {isActive}
            on:click={() => {
              dispatch("select-item", label);
            }}
            color={isActive
              ? "bg-blue-200"
              : activeValues.length
              ? "bg-gray-200"
              : "bg-gray-200"}
          >
            <!-- 
                    title element
                    -------------
                    We will fix the maximum width of the title element
                    to be the container width - pads - the largest value of the right hand.
                    This is somewhat inelegant, but it's a lot more elegant than rewriting the
                    BarAndNumber component to do things that are harder to maintain.
                    The current approach does a decent enough job of maintaining the flow and scan-friendliness.
                 -->
            <div
              class:text-gray-700={!atLeastOneActive}
              class:text-gray-500={atLeastOneActive && !isActive}
              class:italic={atLeastOneActive && !isActive}
              class="w-full text-ellipsis overflow-hidden whitespace-nowrap"
              slot="title"
            >
              {label}
            </div>

            <!-- right-hand metric value -->
            <div slot="right" bind:this={righthandElements[i]}>
              {#if !(atLeastOneActive && !isActive)}
                <div in:fly={{ duration: 200, y: 4 }}>
                  {value}
                </div>
              {/if}
            </div>
          </LeaderboardListItem>
          <TooltipContent slot="tooltip-content">
            <div style:max-width="480px">
              <div>
                {formatBigNumberPercentage(value / referenceValue)} of records
              </div>
              <div>
                filter on <span class="italic">{label}</span>
              </div>
            </div>
          </TooltipContent>
        </Tooltip>
      </div>
    {/each}
    {#if values.length > slice}
      <Tooltip location="right">
        <LeaderboardListItem value={0} color="bg-gray-100">
          <div class="italic text-gray-500" slot="title">All Others</div>
          <div class="italic text-gray-500" slot="right">
            {referenceValue -
              values
                .slice(0, !seeMore ? slice : seeMoreSlice)
                .reduce((a, b) => a + b.value, 0)}
          </div>
        </LeaderboardListItem>
        <TooltipContent slot="tooltip-content">see next 12</TooltipContent>
      </Tooltip>

      <!-- <button
        class="italic pl-2 pr-2 p-1 text-gray-500 w-full text-left hover:bg-gray-50"
        on:click={() => {
          seeMore = !seeMore;
        }}
      >
        {#if seeMore}
          show only top {slice}
        {:else}
          show {seeMoreSlice - slice} more
          {#if hiddenSelectedValues.length}
            ({hiddenSelectedValues.length}
            selected value{#if hiddenSelectedValues.length !== 1}s{/if} hidden.)
          {/if}
        {/if}
      </button> -->
    {/if}
  </LeaderboardList>
</LeaderboardContainer>
