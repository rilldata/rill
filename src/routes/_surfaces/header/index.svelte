<script lang="ts">
import { getContext } from "svelte";
import type { AppStore } from "$lib/app-store"
import {dataModelerService} from "$lib/app-store";

import ModelIcon from "$lib/components/icons/Code.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import EditIcon from "$lib/components/icons/EditIcon.svelte";

const store = getContext('rill:app:store') as AppStore;

function formatModelName(str) {
    let output = str.trim().replaceAll(' ', '_');
    return output;
}

let currentModel;
$: if ($store?.models && $store?.activeAsset) currentModel = $store.models.find(q => q.id === $store.activeAsset.id);

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
    class="grid items-center content-stretch bg-gray-100" 
    style:grid-template-columns="[title] auto [controls] auto">
    <div>
        {#if titleInput !== undefined && titleInput !== null}
        <h1 style:font-size='16px' class="grid grid-flow-col justify-start items-center gap-x-1 p-6">
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