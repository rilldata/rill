<script lang="ts">
    import { format } from "d3-format";
    import BarAndLabel from "$lib/components/BarAndLabel.svelte";
    const formatPercentage = format('.1%');
    const formatCount = format(',');

    export let displaySize:string = "md";
    export let totalRows:number;
    export let topK:any; // FIXME
</script>

<div style="w-full">
    <div class='grid w-full' style="grid-template-columns: auto max-content; justify-items: stretch; justify-content: stretch; grid-column-gap: 1rem;">
        {#each topK.slice(0, 5) as { value, count}}

                <div
                    class="text-gray-500 italic text-ellipsis overflow-hidden whitespace-nowrap {displaySize}-top-k"
                >
                    {value} {value === null ? '∅' : ''}
                </div>
                <BarAndLabel value={count / totalRows}>
                    {formatCount(count)} ({formatPercentage(count / totalRows)})
                </BarAndLabel>
<!-- 
                <div class="text-right" style="position:relative;">
                    {formatCount(count)} ({formatPercentage(count / totalRows)})
                    <div class='number-bar' style="--width: {count / totalRows};" />
                </div> -->
        {/each}
    </div>
</div>

<style>
.number-bar {
    --width: 0%;
    content: '';
    display: inline-block;
    width: calc(100% * var(--width));
    position: absolute;
    left: 0;
    z-index: -1;
    height: 1rem;
    background-color: hsl(280, 50%, 90%);
}
</style>