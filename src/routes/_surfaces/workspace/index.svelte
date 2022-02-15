<script lang="ts">
import { getContext } from "svelte";
import type { AppStore } from '$lib/app-store';
import DropZone from "$lib/components/DropZone.svelte";
import ModelView from "./Model.svelte";
import MetricsDefinitionView from "./MetricsDefinition.svelte";
import ExploreView from "./Explore.svelte";
import {dataModelerService} from "$lib/app-store";
const store = getContext("rill:app:store") as AppStore;
</script>

<!-- <button 
    class="grid justify-end w-full p-3"
    on:click={() => {
        store.action('unsetActiveAsset');
    }} style="font-size:12px;">✕</button> -->

{#if $store?.activeAsset?.assetType === 'model'}
    <ModelView />
{:else if $store?.activeAsset?.assetType === 'metricsDefinition'}
    <MetricsDefinitionView />
{:else if $store?.activeAsset?.assetType === 'exploreConfiguration'}
    <ExploreView />
{:else}
<DropZone 
    padTop={!!$store?.queries?.length}
    on:source-drop={(evt) => { 
    dataModelerService.dispatch('addModel', [{ query: evt.detail.props.content, makeActive: true } ]);
    }} />
{/if}
