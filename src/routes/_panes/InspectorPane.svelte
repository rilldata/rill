<!-- <script>
import { getContext } from "svelte";
import ModelInspectorPane from "./ModelInspectorPane.svelte";

const store = getContext('rill:app:store');
$: activeAsset = $store?.activeAsset;
</script>

{#if activeAsset?.assetType === 'model'}
    <ModelInspectorPane />
{/if}
 -->


 <script lang="ts">
    import { getContext } from "svelte";
    import ModelInspectorPane from "./ModelInspectorPane.svelte";
    
    import type { AppStore } from '$lib/app-store';
    
    import { drag } from "$lib/drag";    
    const store = getContext('rill:app:store') as AppStore;
    
    </script>
        
    <div class='drawer-container flex'>    
      <div class='drawer-handler w-4 absolute hover:cursor-col-resize -translate-x-2 body-height'
      use:drag={{ minSize: 400 }} />
    
      <div class='inspector'>
            {#if $store?.activeAsset?.assetType === 'model'}
            <ModelInspectorPane />
          {/if}
    </div>
    </div>
    <style lang="postcss">
    
    .drawer-container {
      height: calc(100vh - var(--header-height));
    }
    
    .inspector {
      width: var(--right-sidebar-width, 400px);
      font-size: 12px;
    }
    
    </style>