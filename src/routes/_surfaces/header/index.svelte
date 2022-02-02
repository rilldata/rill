<script lang="ts">
import { getContext, createEventDispatcher } from "svelte";
import type { AppStore } from "$lib/app-store"
import ModelIcon from "$lib/components/icons/Code.svelte";
import EditIcon from "$lib/components/icons/EditIcon.svelte";
const dispatch = createEventDispatcher();
const store = getContext('rill:app:store') as AppStore;

function formatModelName(str) {
    let output = str.trim().replaceAll(' ', '_');
    if (!output.endsWith('.sql')) {
        output += '.sql';
    }
    return output;
}

let currentModel;
let sources;
$: if ($store?.queries && $store?.activeAsset) currentModel = $store.queries.find(q => q.id === $store.activeAsset.id);
$: if (currentModel?.sources) sources = $store.sources.filter(source => currentModel.sources.includes(source.path));

let editingTitle = false;
let titleInputValue;
$: titleInput = currentModel?.name;
</script>

<header 
    style:font-size='12px'
    style:height="var(--header-height)" 
    class="grid items-center content-stretch" 
    style:grid-template-columns="[title] auto [controls] auto">
    <div>
        <h1 title="model: {titleInput}" style:font-size='16px' class="grid grid-flow-col justify-start items-center gap-x-3 p-3 pl-5 pr-5">
            <ModelIcon />
            {#if titleInput !== undefined}
                    <input 
                    bind:this={titleInput} 
                    on:input={(evt) => {
                        titleInputValue = evt.target.value;
                        editingTitle = true;
                    }}
                    class:font-bold={editingTitle === false}
                    on:blur={()  => { editingTitle = false; }}
                    value={titleInput} 
                    size={Math.max((editingTitle ? titleInputValue : titleInput)?.length || 0, 5) + 1} 
                    on:change={(e) => { 
                        if (currentModel?.id) {
                            store.action('changeQueryName', { id: currentModel?.id, name: formatModelName(e.target.value) });
                        }
                        
                    }} />
                    <!-- <button class='small-action-button edit-button' on:click={() => {
                        titleInput.focus();
                    }}>
                        <EditIcon />
                    </button> -->
            {/if}
            <!-- {currentModel?.name || ''} -->
        </h1>
    </div>
    <!-- <div class="grid justify-end items-center grid-flow-col pr-5 gap-x-3">
        {#if sources}
        <div class="w-full flex align-items-stretch flex-col">
          <button class="p-2 pt-1 pb-1 bg-transparent bg-black text-white border border-black rounded-md" on:click={() => {
            const query = currentModel.query;
            const exportFilename = currentModel.name.replace('.sql', '.parquet');
            const path = `./export/${exportFilename}`
            store.action('exportToParquet', {query, path, id: currentModel.id });
          }}>generate {currentModel.name.replace('.sql', '.parquet')}</button>
        </div>
      {/if}
    </div> -->
</header>