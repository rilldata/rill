<script lang="ts">
import IconButton from "$lib/components/IconButton.svelte";
import * as classes from "$lib/util/component-classes";

import { defaultSort, sortByNullity, sortByName } from "$lib/components/collapsible-table-summary/sort-utils";
export let sortMethod = defaultSort;
export let previewView = 'summaries';

</script>

<!-- <div class='grid grid-flow-col justify-between'>
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
</div> -->

<div class='pt-2 pb-2 flex justify-between text-gray-500'>
    <select bind:value={sortMethod} class={classes.NATIVE_SELECT}>
        <!-- <option value={sortByOriginalOrder}>sort by original order</option> -->
        <option value={defaultSort}>sort by data type</option>
        <option value={sortByNullity}>sort by null %</option>
        <option value={sortByName}>sort by name</option>
    </select>
    <select bind:value={previewView} class={classes.NATIVE_SELECT}>
        <option value="summaries">show summaries</option>
        <option value="example">show example</option>
    </select>
</div>