<script lang="ts">
import { getContext, onMount, tick } from "svelte";
import { slide } from "svelte/transition";
import { tweened } from "svelte/motion";
import { sineOut as easing } from "svelte/easing";
import CollapsibleSectionTitle from "$lib/components/CollapsibleSectionTitle.svelte";
import ColumnProfile from "$lib/components/column-profile/ColumnProfile.svelte";
import ContextButton from "$lib/components/column-profile/ContextButton.svelte";
import Spacer from "$lib/components/icons/Spacer.svelte";
import * as classes from "$lib/util/component-classes";
import Export from "$lib/components/icons/Export.svelte";
import { onClickOutside } from "$lib/util/on-click-outside";
import Menu from "$lib/components/menu/Menu.svelte"
import MenuItem from "$lib/components/menu/MenuItem.svelte"

import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";

import type { ApplicationStore } from "$lib/app-store";

import { formatInteger, formatBigNumberPercentage } from "$lib/util/formatters"

import {dataModelerService} from "$lib/app-store";
import type {
    PersistentTableEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type {
    DerivedTableEntity
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { DerivedTableStore, PersistentTableStore } from "$lib/tableStores";
import type { DerivedModelStore, PersistentModelStore } from "$lib/modelStores";
import FloatingElement from "$lib/components/tooltip/FloatingElement.svelte";
import CollapsibleTableSummary from "$lib/components/column-profile/CollapsibleTableSummary.svelte";

import { config } from "$lib/components/column-profile/utils"

const persistentTableStore = getContext('rill:app:persistent-table-store') as PersistentTableStore;
const derivedTableStore = getContext('rill:app:derived-table-store') as DerivedTableStore;

const store = getContext('rill:app:store') as ApplicationStore;
const queryHighlight = getContext('rill:app:query-highlight');

function tableDestinationCompute(key, table, destination) {
    let inputs = table.reduce((acc,v) => acc + v[key], 0)
    return  (destination[key]) / inputs;
}

function computeRollup(table, destination) {
    return tableDestinationCompute('cardinality', table, destination);
}

let rollup;
let tables;
// get source tables?
let sourceTableReferences;
let showColumns = true;
let showExportOptions = true;
let sourceTableNames = [];

// interface tweens for the  big numbers
let bigRollupNumber = tweened(0, { duration: 700, easing });
let inputRowCardinality = tweened(0, { duration: 200, easing });
let outputRowCardinality = tweened(0, { duration: 250, easing });

let currentTable: PersistentTableEntity;
    $: currentTable = ($store?.activeEntity && $persistentTableStore?.entities) ?
        $persistentTableStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;
    let currentDerivedTable: DerivedTableEntity;
    $: currentDerivedTable = ($store?.activeEntity && $derivedTableStore?.entities) ?
        $derivedTableStore.entities.find(q => q.id === $store.activeEntity.id) : undefined;

// toggle state for inspector sections
let showSourceTables = true;


let container;
let containerWidth = 0;
let contextMenu;
let contextMenuOpen = false;
let menuX;
let menuY;
let clickOutsideListener;
$: if (!contextMenuOpen && clickOutsideListener) {
    clickOutsideListener();
    clickOutsideListener = undefined;
}

onMount(() => {
    const observer = new ResizeObserver(entries => {
        containerWidth = container.clientWidth;
    });
    observer.observe(container);
})
</script>

{#key currentTable?.id}
    <div bind:this={container}>

    <hr />
        <hr />
        
        <div class="pb-4 pt-4">
        <div class=" pl-4 pr-4">
        <CollapsibleSectionTitle tooltipText="source tables" bind:active={showColumns}>
            Table Columns
        </CollapsibleSectionTitle>
        </div>

        {#if currentDerivedTable?.profile}
        <div  transition:slide|local={{duration: 200}}>
            <CollapsibleTableSummary
                showTitle={false}
                showContextButton={false}
                show={showColumns}
                name={currentTable.name}
                cardinality={currentDerivedTable?.cardinality ?? 0}
                profile={currentDerivedTable?.profile ?? []}
                head={currentDerivedTable?.preview ?? []}
                emphasizeTitle ={currentTable?.id === $store?.activeEntity?.id}
            />
        </div>

        {/if}
    </div>

    </div>
{/key}