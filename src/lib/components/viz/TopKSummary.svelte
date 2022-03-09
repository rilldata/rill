<script lang="ts">
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import StackingWord from "$lib/components/tooltip/StackingWord.svelte";
import TooltipShortcutContainer from "$lib/components/tooltip/TooltipShortcutContainer.svelte";
import Shortcut from "$lib/components/tooltip/Shortcut.svelte";
import { format } from "d3-format";
import BarAndLabel from "$lib/components/BarAndLabel.svelte";
import notificationStore from "$lib/components/notifications/";

//const formatPercentage = format('.1%');
const formatCount = format(',');

export let displaySize:string = "md";
export let totalRows:number;
export let topK:any; // FIXME
export let color:string;

$: smallestPercentage = Math.min(...topK.slice(0,5).map(entry => entry.count / totalRows))
$: formatPercentage = smallestPercentage < 0.001 ? 
    format('.2%') : 
    format('.1%');

let shiftClicked = false;
let shiftClickedTimeout;
let CLICK_DURATION = 300;
</script>

<div class='w-full'>
    <div class='grid w-full' style="
        grid-template-columns: auto  max-content; 
        grid-auto-rows: 19px;
        justify-items: stretch; 
        justify-content: stretch; 
        grid-column-gap: 1rem;"
    >
        {#each topK.slice(0, 10) as { value, count}}
            {@const printValue = value === null ? ' null ∅' : value }
                <Tooltip location="right" alignment="center">
                <div class="text-gray-500 italic text-ellipsis overflow-hidden whitespace-nowrap {displaySize}-top-k"

                on:click={async (event) => {
                    // we should only allow activation when there are rows present.
                    if (event.shiftKey) {
                        await navigator.clipboard.writeText(value);
            
                        notificationStore.send({ message: `copied column value "${value === null ? 'NULL' : value}" to clipboard`});
                        clearTimeout(shiftClickedTimeout);
                        shiftClicked = true;
                        shiftClickedTimeout = setTimeout(() => {
                            shiftClicked = false;
                        }, CLICK_DURATION);
                    }
                }}
                >
                        {printValue}
                </div>
                <TooltipContent slot="tooltip-content">
                    <div>
                    {printValue}
                    </div>
                    <TooltipShortcutContainer>
                        <div>
                            <StackingWord active={shiftClicked}>copy</StackingWord> column value to clipboard
                        </div>
                        <Shortcut>
                            shift + click
                        </Shortcut>
                    </TooltipShortcutContainer>
                </TooltipContent>
                </Tooltip>
                {@const negligiblePercentage = count / totalRows < .001}
                {@const percentage = negligiblePercentage ? 'ϵ%' : formatPercentage(count / totalRows)}
                <BarAndLabel value={count / totalRows} {color}>
                    <span class:text-gray-500={negligiblePercentage}>{formatCount(count)} ({percentage})</span>
                </BarAndLabel>
        {/each}
    </div>
</div>