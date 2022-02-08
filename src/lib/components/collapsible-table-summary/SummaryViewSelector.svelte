<script lang="ts">
import IconButton from "$lib/components/IconButton.svelte";
import { defaultSort, sortByNullity, sortByName } from "$lib/components/collapsible-table-summary/sort-utils";
export let sortMethod = defaultSort;
export let previewView = 'card';

</script>

<div class='grid grid-flow-col justify-between'>
    <div class='grid justify-start grid-cols-3 items-baseline pb-2 pt-2'>
        <IconButton title="sort by cardinality" selected={sortMethod === defaultSort} on:click={() => {
            sortMethod = defaultSort;
        }}>||</IconButton>

        <IconButton 
            title="sort by null % of field"
            selected={sortMethod === sortByNullity}  on:click={() => {
            sortMethod = sortByNullity;
        }}>∅</IconButton>
        <IconButton title="sort by name" selected={sortMethod === sortByName}  on:click={() => {
            sortMethod = sortByName;
        }}>AZ</IconButton>
    </div>

    <!-- <div class='grid justify-start grid-cols-2 items-baseline pb-2 pt-2'>
        {#if draggable}
        <IconButton 
            title="select columns"
            selected={selectingColumns} on:click={() => {
            selectingColumns = !selectingColumns;
        }}>*</IconButton>
        {/if}
        <IconButton title="summarize columns" selected={showSummaries} on:click={() => {
            showSummaries = !showSummaries;
        }}>
            &
        </IconButton>
    </div> -->

    <div class='grid justify-end grid-flow-col pb-2 pt-2 '>
        <IconButton 
            title="show summaries (cardinalities, null %, and time range)"
            selected={previewView === 'card'} 
            on:click={() => {
                previewView = 'card';
        }}>||</IconButton>
        <IconButton 
            title="show example value"
            selected={previewView === 'example'}  
            on:click={() => {
                previewView = 'example';
        }}>Ex</IconButton>
    </div>
</div>