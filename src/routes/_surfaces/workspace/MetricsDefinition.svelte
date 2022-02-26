<script lang="ts">
import { getContext } from "svelte";
import { slide } from "svelte/transition";
import { cubicOut as easing } from "svelte/easing";
import type { AppStore } from '$lib/app-store';

import MetricsEditor from "$lib/components/MetricsEditor.svelte";
import {dataModelerService} from "$lib/app-store";

const store = getContext("rill:app:store") as AppStore;

$: currentModel = ($store && $store?.metricsModels) ? $store?.metricsModels.find(model => model.id === $store?.activeAsset?.id) : undefined;
</script>

<section class='bg-gray-100'>
    {#if currentModel}
        {#key currentModel.id}
            <MetricsEditor
                content={currentModel.spec}
                name={currentModel.name}
                on:process={() => {
                    //dataModelerService.dispatch('createExploreConfiguration', [{ metricsModelID: currentModel.id, name: currentModel.name }]);
                }}
                on:rename={(event) => {
                    //dataModelerService.dispatch('updateMetricsModelName', [{ id: currentModel.id, name: event.detail} ])
                }}
                on:delete={() => {
                    //dataModelerService.dispatch('deleteMetricsModel', [{ id: currentModel.id }])
                }}
                on:write={(event) => {
                    const newSpec = event.detail.content;
                    //dataModelerService.dispatch('updateMetricsModelSpec', [{ id: currentModel.id, newSpec }])
                }}
            >
            <svelte:fragment slot='prototype-container'>
                {#if currentModel?.error !== undefined}
                <div transition:slide={{ duration: 200, easing }} 
                  class="error p-4 m-4 rounded-lg shadow-md"
                >{currentModel.error}</div>
                {/if}
                <pre>
                    {JSON.stringify(currentModel.parsedSpec, null, 2)}
                </pre>
            </svelte:fragment>
            </MetricsEditor>
         {/key}
    {/if}

</section>

<style>
section {
    grid-column: 2;
    width: 100%;
}
</style>