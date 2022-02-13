<script>
import { getContext } from "svelte";

import { dragVertical } from "$lib/drag";
import RowTable from "$lib/components/RowTable.svelte";
import PreviewTable from "$lib/components/table/PreviewTable.svelte";

const store = getContext('rill:app:store');
let currentQuery;
$: if ($store?.queries && $store?.activeAsset) currentQuery = $store.queries.find(q => q.id === $store.activeAsset.id);
</script>

<div 
    class=""
    style:width="calc(100vw - var(--left-sidebar-width))"
    style:height="var(--bottom-sidebar-width)"
>

    <div 
        class="relative  bg-gray-50 overflow-x-auto border-t border-gray-300" 
        style:height="var(--bottom-sidebar-width)"
        style:min-width="calc(100vw - var(--left-sidebar-width) - .5rem)"
        style:font-size="12px">
        {#if currentQuery?.preview}
        <PreviewTable rows={currentQuery.preview} columnNames={currentQuery.profile} />
        <!-- <RowTable data={currentQuery.preview} /> -->
    {/if}
    </div>

    <div 
    class='drawer-handler h-4 absolute hover:cursor-col-resize translate-y-3 z-20'
    style:bottom="var(--bottom-sidebar-width, 300px)"
    style:min-width="300px"
    style:width="calc(100vw - var(--left-sidebar-width) - .5rem)"
    use:dragVertical={{minSize: 0}}
/>
</div>