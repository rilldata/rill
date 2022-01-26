<script>
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
                    class:italic={!model.spec.length && model.id !== $store.activeAsset.id} 
                    class:font-bold={model.id === $store.activeAsset.id} 
                    on:click={() => { store.action('setActiveAsset', {id: model.id, assetType: 'metricsModel'}) }}>
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