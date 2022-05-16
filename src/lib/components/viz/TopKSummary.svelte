<script lang="ts">
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
import { format } from "d3-format";
import BarAndLabel from "$lib/components/BarAndLabel.svelte";
import notificationStore from "$lib/components/notifications/";
import { config } from "$lib/components/column-profile/utils";

import { createShiftClickAction } from "$lib/util/shift-click-action";

export let displaySize:string = "md";
export let totalRows:number;
export let topK:any; // FIXME
export let color:string;
export let containerWidth:number;

const { shiftClickAction } = createShiftClickAction();

$: smallestPercentage = Math.min(...topK.slice(0,5).map(entry => entry.count / totalRows))
$: formatPercentage = smallestPercentage < 0.01 ? 
    format('0.2%') : 
    format('0.1%');

$: formatCount = format(',');

// time to create a single way to get the width of an element.

</script>

<div class='w-full select-none'>
    <div class='grid w-full' style="
        grid-template-columns: auto  max-content; 
        grid-auto-rows: 19px;
        justify-items: stretch; 
        justify-content: stretch; 
        grid-column-gap: 1rem;"
    >
        {#each topK.slice(0, 10) as { value, count}}
            {@const printValue = value === null ? ' null ∅' : value }
                <Tooltip location="right" alignment="center"  distance={16}>
                <div class="text-gray-500 italic text-ellipsis overflow-hidden whitespace-nowrap {displaySize}-top-k"
                use:shiftClickAction
                on:shift-click={async () => {
                    await navigator.clipboard.writeText(value);
                    notificationStore.send({ message: `copied column value "${value === null ? 'NULL' : value}" to clipboard`});
                }}
                >
                        {printValue}
                </div>
                <TooltipContent slot="tooltip-content">
                    <div class="pt-1 pb-1 italic" style:max-width="360px">
                        {printValue}
                    </div>
                    <TooltipShortcutContainer>
                        <div>
                            <StackingWord>copy</StackingWord> column value to clipboard
                        </div>
                        <Shortcut>
                            <span style='font-family: var(--system);'>⇧</span> + Click
                        </Shortcut>
                    </TooltipShortcutContainer>
                </TooltipContent>
                </Tooltip>
                {@const negligiblePercentage = count / totalRows < .0002}
                {@const percentage = negligiblePercentage ? '<.01%' : formatPercentage(count / totalRows)}
                <Tooltip location="right" alignment="center" distance={16}>
                    <div
                    use:shiftClickAction
                    on:shift-click={async () => {
                        await navigator.clipboard.writeText(count);
                        notificationStore.send({ message: `copied column value "${count === null ? 'NULL' : count}" to clipboard`});
                    }}
                    >
                    <BarAndLabel value={count / totalRows} {color}>
                        <span class:text-gray-500={negligiblePercentage && containerWidth >= config.hideRight}>{formatCount(count)} 
                            {#if (!containerWidth) || containerWidth >= config.hideRight}
                                {#if percentage.length < 6}&nbsp;{/if}{#if percentage.length < 5}&nbsp;{/if}&nbsp;<span class:text-gray-600={!negligiblePercentage}>({percentage})</span>
                            {/if}
                        </span>
                    </BarAndLabel>
                </div>
                   
                    <TooltipContent slot='tooltip-content'>
                        <div class="pt-1 pb-1 italic" style:max-width="360px">
                            {formatCount(count)} ({percentage})
                        </div>
                        <TooltipShortcutContainer>
                            <div>
                                <StackingWord>copy</StackingWord> {count} to clipboard
                            </div>
                            <Shortcut>
                                <span style='font-family: var(--system);'>⇧</span> + Click
                            </Shortcut>
                        </TooltipShortcutContainer>
                    </TooltipContent>
                </Tooltip>
        {/each}
    </div>
</div>