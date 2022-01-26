<script>
import Model from "$lib/components/icons/Model.svelte";
import { getContext } from "svelte";
const store = getContext("rill:app:store");
</script>
<section class='p-5'>
    <button on:click={() => {
        store.action('createMetricsModel');
    }}>add model</button>
    {#if $store?.metricsModels.length}
        {#each $store.metricsModels as model, i (model)}
            <div>
                <button
                    class="grid" 
                    class:italic={!model.spec.length && model.id !== $store.activeMetricsModel} 
                    class:font-bold={model.id === $store.activeMetricsModel} 
                    on:click={() => { store.action('setActiveMetricsModel', {id: model.id}) }}>
                    {model.name}
                </button>
            </div>
        {/each}
    {/if}
</section>

<style>
section {
    font-size:12px;
    min-width: 300px;
    grid-column: 1;
}
    
</style>