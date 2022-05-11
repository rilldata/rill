<script lang="ts">
import { createEventDispatcher } from "svelte";
import { fly } from 'svelte/transition';

import LeaderboardContainer from "$lib/components/leaderboard/LeaderboardContainer.svelte";
import LeaderboardHeader from "$lib/components/leaderboard/LeaderboardHeader.svelte";
import LeaderboardList from "$lib/components/leaderboard/LeaderboardList.svelte";
import LeaderboardListItem from "$lib/components/leaderboard/LeaderboardListItem.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

export let displayName;
export let total:number;
export let values;
export let nullCount = 0;
export let activeValues:string[];

export let slice = 3;
export let seeMoreSlice = 15;

let seeMore = false;

const dispatch = createEventDispatcher();

$: atLeastOneActive = activeValues?.length || 0;

</script>
<LeaderboardContainer>
    <Tooltip location="top" alignment="start" distance={4}>
    <LeaderboardHeader isActive={atLeastOneActive}>
        <div slot="title">{displayName}</div>
        <div slot="right">
            {#if activeValues.length}
                <button 
                    in:fly|local={{ duration: 100, y: 4 }} 
                    out:fly|local={{ duration: 100, y: -4}}
                    on:click={() => { dispatch('clear-all')}}>clear</button>
            {/if}
        </div>
    </LeaderboardHeader>
    <TooltipContent slot="tooltip-content">
        hmm
    </TooltipContent>
    </Tooltip>
    <LeaderboardList>
        {#each values.slice(0, !seeMore ? slice : seeMoreSlice) as {label, value} (label)}
            {@const isActive = activeValues?.includes(label)}
        <div>
        <Tooltip location="right">
            <LeaderboardListItem
                value={atLeastOneActive && !isActive ? 0 : value / total}
                {isActive}
                on:click={() => { dispatch('select-item', label) }}
                color={atLeastOneActive && !isActive ? 'bg-gray-200' : 'bg-blue-200'}
            >
                <div
                    class:text-gray-500={atLeastOneActive && !isActive}
                           class:italic={atLeastOneActive && !isActive}
                    class="w-full text-ellipsis overflow-hidden whitespace-nowrap" 
                    slot="title">
                    {label}
                </div>
                <div slot="right">
                    {#if !(atLeastOneActive && !isActive)}

                        <div in:fly={{duration: 200, y: 4}}>
                            {value}
                        </div>
                    {/if}
                </div>
            </LeaderboardListItem>
            <TooltipContent slot='tooltip-content'>
                <div>
                    {value / total}
                </div>
                <div>
                    filter on <span class='italic'>{label}</span>
                </div>
            </TooltipContent>
            </Tooltip>
        </div>
        {/each}
        {#if nullCount}
        <LeaderboardListItem
            value={nullCount / total}
            color="bg-gray-100"
        >
            <div class="italic text-gray-500" slot="title">
                nulls
            </div>
            <div class="italic text-gray-500" slot="right">  {nullCount}</div>    
        </LeaderboardListItem>
        {/if}
        <hr />
        {#if !seeMore}
            <Tooltip location="right">
            <LeaderboardListItem
                value={1 - (values.slice(0, slice).reduce((a,b) => a+b.value, 0)) / total}
                color="bg-gray-100"
                on:click={() => {
                    seeMore = true;
                }}
            >
                <div class="italic text-gray-500" slot="title">
                    Other
                </div>
                <div class="italic text-gray-500" slot="right">{total - values.slice(0,slice).reduce((a,b) => a + b.value, 0) - nullCount}</div>    
            </LeaderboardListItem>
            <TooltipContent slot="tooltip-content">
                see next 12
            </TooltipContent>
            </Tooltip>
        {:else}
            <button 
                class="italic pl-2 pr-2 p-1 text-gray-500"
                on:click={() => {
                    seeMore = false;
                }}>show only top {slice}</button>
        {/if}
    </LeaderboardList>
</LeaderboardContainer>