<script lang="ts">
import IconButton from "$lib/components/IconButton.svelte";
import * as classes from "$lib/util/component-classes";

import { defaultSort, sortByNullity, sortByName } from "$lib/components/collapsible-table-summary/sort-utils";
export let sortMethod = defaultSort;
export let previewView = 'summaries';

export let compact = false;

</script>

<div class='pt-2 pb-2 pl-14 flex justify-between text-gray-500'>
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