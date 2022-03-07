<script lang="ts">
import { getContext } from "svelte";
import type { ApplicationStore } from "$lib/app-store";
import DropZone from "$lib/components/DropZone.svelte";
import ModelView from "./Model.svelte";
// import MetricsDefinitionView from "./MetricsDefinition.svelte";
// import ExploreView from "./Explore.svelte";
import {dataModelerService} from "$lib/app-store";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { PersistentModelStore } from "$lib/modelStores";
const store = getContext("rill:app:store") as ApplicationStore;
const persistentModelStore = getContext('rill:app:persistent-model-store') as PersistentModelStore;
</script>

{#if $store?.activeEntity?.type === EntityType.Model}
    <ModelView />
{:else}
<DropZone
    padTop={!!$persistentModelStore?.entities.length}
    on:source-drop={(evt) => {
    dataModelerService.dispatch('addModel', [{ query: evt.detail.props.content, makeActive: true } ]);
    }} />
{/if}
