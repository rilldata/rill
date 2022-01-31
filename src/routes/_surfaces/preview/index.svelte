<script>
import { getContext } from "svelte";

import { dragVertical } from "$lib/drag";
import RowTable from "$lib/components/RowTable.svelte";


const store = getContext('rill:app:store');
let currentQuery;
$: if ($store?.queries && $store?.activeAsset) currentQuery = $store.queries.find(q => q.id === $store.activeAsset.id);
</script>

<div class="border-t border-gray-300 p-3">
    <div 
        class='drawer-handler w-full h-4 absolute hover:cursor-col-resize -translate-y-5'
        use:dragVertical={{minSize: 0}}
    />
    {#if currentQuery?.preview}
        <RowTable data={currentQuery.preview} />
    {/if}
</div>