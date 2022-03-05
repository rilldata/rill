<script lang="ts">
import { getContext } from "svelte";
import { ApplicationStore, dataModelerService } from "$lib/app-store";

import ModelIcon from "$lib/components/icons/Code.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import EditIcon from "$lib/components/icons/EditIcon.svelte";
import { PersistentModelStore } from "$lib/modelStores";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";

const store = getContext('rill:app:store') as ApplicationStore;
const persistentModelStore = getContext('rill:app:persistent-model-store') as PersistentModelStore;

function formatModelName(str) {
    let output = str.trim().replaceAll(' ', '_');
    return output;
}

let currentModel: PersistentModelEntity;
$: if ($store?.activeEntity && $persistentModelStore?.entities)
    currentModel = $persistentModelStore.entities.find(q => q.id === $store.activeEntity.id);

let titleInputElement;
let editingTitle = false;
let titleInputValue;
let tooltipActive;
let titleInput = currentModel?.name;
$: titleInput = currentModel?.name;


function onKeydown(event) {
    if (editingTitle && event.key === 'Enter') {
        titleInputElement.blur();
    }
}
</script>

<svelte:window on:keydown={onKeydown} />

<header 
    style:font-size='12px'
    style:height="var(--header-height)"
    class="grid items-center content-stretch bg-gray-100 pl-6 pr-6" 
    style:grid-template-columns="[title] auto [controls] auto">
    <div>
        {#if titleInput !== undefined && titleInput !== null}
        <h1 style:font-size="16px"  class="grid grid-flow-col justify-start items-center gap-x-1">
            <ModelIcon />
            <Tooltip distance={8} bind:active={tooltipActive} suppress={editingTitle}>
                <input 
                bind:this={titleInputElement} 
                on:input={(evt) => {
                    titleInputValue = evt.target.value;
                    editingTitle = true;
                }}

                class="bg-gray-100 border border-transparent border-2 hover:border-gray-400 rounded pl-2 pr-2 cursor-pointer"
                class:font-bold={editingTitle === false}
                on:blur={() => { editingTitle = false; }}
                value={titleInput} 
                size={Math.max((editingTitle ? titleInputValue : titleInput)?.length || 0, 5) + 1} 
                on:change={(e) => { 
                    if (currentModel?.id) {
                        dataModelerService.dispatch('updateModelName', [currentModel?.id, formatModelName(e.target.value)]);
                    }
                }} />
            <TooltipContent slot="tooltip-content">
                <div class='flex items-center'><EditIcon size=".75em"} />Edit</div>
            </TooltipContent>
            </Tooltip>
        </h1>
        {/if}
    </div>
</header>