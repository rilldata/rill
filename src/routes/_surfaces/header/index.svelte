<script lang="ts">
import { getContext } from "svelte";
import type { AppStore } from "$lib/app-store"
import ModelIcon from "$lib/components/icons/Code.svelte";
import {dataModelerService} from "$lib/app-store";
const store = getContext('rill:app:store') as AppStore;

function formatModelName(str) {
    let output = str.trim().replaceAll(' ', '_');
    return output;
}

let currentModel;
$: if ($store?.models && $store?.activeAsset) currentModel = $store.models.find(q => q.id === $store.activeAsset.id);

let editingTitle = false;
let titleInputValue;
$: titleInput = currentModel?.name;
</script>

<header 
    style:font-size='12px'
    style:height="var(--header-height)" 
    class="grid items-center content-stretch bg-gray-100" 
    style:grid-template-columns="[title] auto [controls] auto">
    <div>
        {#if titleInput !== undefined && titleInput !== null}
        <h1 title="model: {titleInput}" style:font-size='16px' class="grid grid-flow-col justify-start items-center gap-x-3 p-6">
            <ModelIcon />
                <input 
                bind:this={titleInput} 
                on:input={(evt) => {
                    titleInputValue = evt.target.value;
                    editingTitle = true;
                }}
                class:font-bold={editingTitle === false}
                class="bg-gray-100"
                on:blur={() => { editingTitle = false; }}
                value={titleInput} 
                size={Math.max((editingTitle ? titleInputValue : titleInput)?.length || 0, 5) + 1} 
                on:change={(e) => { 
                    if (currentModel?.id) {
                        dataModelerService.dispatch('updateModelName', [currentModel?.id, formatModelName(e.target.value)]);
                    }
                }} />
        </h1>
        {/if}
    </div>
</header>