<script>
import { getContext } from "svelte";
import ModelInspectorPane from "./ModelInspectorPane.svelte";

const store = getContext('rill:app:store');
$: activeAsset = $store?.activeAsset;
</script>

{#if activeAsset?.assetType === 'model'}
    <ModelInspectorPane />
{/if}

