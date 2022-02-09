<script lang="ts">
    import { format } from "d3-format";
    import BarAndLabel from "$lib/components/BarAndLabel.svelte";
    //const formatPercentage = format('.1%');
    const formatCount = format(',');

    export let displaySize:string = "md";
    export let totalRows:number;
    export let topK:any; // FIXME

    $: smallestPercentage = Math.min(...topK.slice(0,5).map(entry => entry.count / totalRows))
    $: formatPercentage = smallestPercentage < 0.001 ? 
        format('.2%') : 
        format('.1%');

</script>

<div>
    <div class='grid w-full' style="
        grid-template-columns: auto  max-content; 
        grid-auto-rows: 19px;
        justify-items: stretch; 
        justify-content: stretch; 
        grid-column-gap: 1rem;"
    >
        {#each topK.slice(0, 5) as { value, count}}
                <div
                    class="text-gray-500 italic text-ellipsis overflow-hidden whitespace-nowrap {displaySize}-top-k"
                >
                    {value} {value === null ? '∅' : ''}
                </div>
                {@const negligiblePercentage = count / totalRows < .001}
                {@const percentage = negligiblePercentage ? 'ϵ%' : formatPercentage(count / totalRows)}
                <BarAndLabel value={count / totalRows} color='hsl(340,50%, 87%)'>
                    <span class:text-gray-500={negligiblePercentage}>{formatCount(count)} ({percentage})</span>
                </BarAndLabel>
        {/each}
    </div>
</div>
<!-- 
<div style="w-full">
    <table>
        {#each topK.slice(0, 5) as { value, count}}
            <tr>
                <td
                    class="text-gray-500 italic text-ellipsis overflow-hidden whitespace-nowrap {displaySize}-top-k"
                >
                    {value} {value === null ? '∅' : ''}
            </td>
            <td>
                <BarAndLabel value={count / totalRows}>
                    {formatCount(count)} ({formatPercentage(count / totalRows)})
                </BarAndLabel>
            </td>
            </tr>
        {/each}
    </table>
</div> -->