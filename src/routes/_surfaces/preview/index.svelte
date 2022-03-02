<script lang="ts">
import { getContext } from "svelte";

import { dragVertical } from "$lib/drag";
import PreviewTable from "$lib/components/table/PreviewTable.svelte";
import { ApplicationStore } from "$lib/app-store";
import { DerivedModelStore } from "$lib/modelStores";
import type {
    DerivedModelEntity
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";

const store = getContext('rill:app:store') as ApplicationStore;
const derivedModelStore = getContext('rill:app:derived-model-store') as DerivedModelStore;
let currentModel: DerivedModelEntity;
$: if ($store?.activeEntity && $derivedModelStore?.entities)
    currentModel = $derivedModelStore?.entities.find(q => q.id === $store.activeEntity.id);
</script>

<div 
    style:width="calc(100vw - var(--left-sidebar-width))"
    style:height="var(--bottom-sidebar-width)"
>

    <div 
        class="relative  bg-gray-50 overflow-auto border-t border-gray-300" 
        style:height="var(--bottom-sidebar-width)"
        style:min-width="calc(100vw - var(--left-sidebar-width) - .5rem)"
        style:font-size="12px">
        {#if currentModel?.preview}
        <!-- <RowTable data={currentQuery.preview} /> -->
        <PreviewTable rows={currentModel.preview} columnNames={currentModel.profile} />
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