<script lang="ts">
import { getContext, tick } from "svelte";
import { ApplicationStore, dataModelerService } from "$lib/app-store";

import ModelIcon from "$lib/components/icons/Code.svelte";
import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
import EditIcon from "$lib/components/icons/EditIcon.svelte";
import MoreHorizontal from "$lib/components/icons/MoreHorizontal.svelte";
import Export from "$lib/components/icons/Export.svelte"
import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte"
import Menu from "$lib/components/menu/Menu.svelte"
import MenuItem from "$lib/components/menu/MenuItem.svelte"
import type { PersistentModelStore } from "$lib/modelStores";

import { onClickOutside } from "$lib/util/on-click-outside";

import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import Spinner from "$lib/components/Spinner.svelte";

const store = getContext('rill:app:store') as ApplicationStore;
const persistentModelStore = getContext('rill:app:persistent-model-store') as PersistentModelStore;

function formatModelName(str) {
    let output = str
        .trim()
        .replaceAll(' ', '_')
        .replace(/\.sql/, '');
    return output;
}

let currentModel: PersistentModelEntity;
$: if ($store?.activeEntity && $persistentModelStore?.entities)
    currentModel = $persistentModelStore.entities.find(q => q.id === $store.activeEntity.id);

let titleInputElement;
let editingTitle = false;
let titleInputValue;
let tooltipActive;

let menuX;
let menuY;
let clickOutsideListener;

$: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
}

let titleInput = currentModel?.name;
$: titleInput = currentModel?.name;


function onKeydown(event) {
    if (editingTitle && event.key === 'Enter') {
        titleInputElement.blur();
    }
}

let contextMenu;
let contextMenuOpen = false;

// debounce the application status to resolve any quick flickers that
// may occur from quick changes to the application status.
let applicationStatus = 0;
let asTimer;
function debounceStatus(status:EntityStatus) {
    clearTimeout(asTimer);
    asTimer = setTimeout(() => {
        applicationStatus = status;
    }, 100);

}

$: debounceStatus(($store?.status as unknown) as EntityStatus);

</script>

<svelte:window on:keydown={onKeydown} />

<header 
    style:height="var(--header-height)"
    class="grid items-center content-stretch justify-between bg-gray-100 pl-6 pr-6" 
    style:grid-template-columns="[title] auto [controls] auto">
    <div>
        {#if titleInput !== undefined && titleInput !== null}
        <h1 style:font-size="16px"  class="grid grid-flow-col justify-start items-center gap-x-1">
            <ModelIcon />
            <Tooltip distance={8} bind:active={tooltipActive} suppress={editingTitle}>
                <input 
                id="model-title-input"
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
                <div class='flex items-center'><EditIcon size=".75em" />Edit</div>
            </TooltipContent>
            </Tooltip>
        </h1>
        {/if}
    </div>
    <div>
    <div class="text-gray-400">
        <Tooltip>
            <Spinner status={applicationStatus || EntityStatus.Idle} size="24px" />
        <TooltipContent slot="tooltip-content">
            {#if applicationStatus === EntityStatus.Idle}
                idle
            {:else if applicationStatus === EntityStatus.Running}
                running
            {:else if applicationStatus === EntityStatus.Exporting}
                exporting a model resultset
            {:else if applicationStatus === EntityStatus.Importing}
                importing a table
            {:else if applicationStatus === EntityStatus.Profiling}
                profiling
            {/if}
        </TooltipContent>
        </Tooltip>
    </div>
    
<!-- <Tooltip location="left" alignment="middle" distance={16} suppress={contextMenuOpen}>
    <button
    bind:this={contextMenu}
    on:click={async (event) => {
        contextMenuOpen = !contextMenuOpen;
        menuX = event.clientX;
        menuY = event.clientY;

        if (!clickOutsideListener) {
            await tick();
            clickOutsideListener = onClickOutside(() => {
                contextMenuOpen = false;
            }, contextMenu);
        }
    }}
    style:grid-column="left-control"
    class="
        hover:bg-gray-300
        transition-tranform 
        text-gray-500
        duration-100
        items-center
        justify-center
        border
        border-transparent
        rounded
        flex flex-row gap-x-2
        pl-2 pr-2
        pt-1 pb-1
       "
    >
    export
    <Export size="16px" />
</button>
    <TooltipContent slot="tooltip-content">
        export this model as a dataset
    </TooltipContent>
</Tooltip> -->

{#if contextMenuOpen}
<!-- place this above codemirror.-->
<div bind:this={contextMenu}>
    <FloatingElement relationship="mouse" target={{x: menuX, y:menuY}} location="left" alignment="start">
        <Menu on:escape={()=> { contextMenuOpen = false; }} on:item-select={() => { contextMenuOpen = false; }}>
            <MenuItem on:select={() => {
                const exportFilename = currentModel.name.replace('.sql', '.parquet');
                dataModelerService.dispatch('exportToParquet', [currentModel.id, exportFilename]);
            }}>
                Export as Parquet 
            </MenuItem>
            <MenuItem on:select={() => {
                const exportFilename = currentModel.name.replace('.sql', '.csv');
                dataModelerService.dispatch('exportToCsv', [currentModel.id, exportFilename]);
            }}>
                Export as CSV 
            </MenuItem>
        </Menu>
    </FloatingElement>
</div>
{/if}

    </div>
</header>