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
import ModelInspector from "./Model.svelte";

import type { ApplicationStore } from "$lib/app-store";

import { drag } from "$lib/drag";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
const store = getContext('rill:app:store') as ApplicationStore;

</script>
        
  <div class='drawer-container flex'>    
    <div class='drawer-handler w-4 absolute hover:cursor-col-resize -translate-x-2 body-height'
    use:drag={{ minSize: 400 }} />
  
    <div class='inspector'>
          {#if $store?.activeEntity?.type === EntityType.Model}
          <ModelInspector />
        {/if}
  </div>
  </div>
  <style lang="postcss">
  
  .inspector {
    width: var(--right-sidebar-width, 400px);
    font-size: 12px;
  }
  
  </style>