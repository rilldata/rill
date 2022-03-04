<script lang="ts">
import { getContext } from "svelte";
import ModelInspector from "./Model.svelte";

import type { ApplicationStore } from "$lib/app-store";

import { drag } from "$lib/drag";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
const store = getContext('rill:app:store') as ApplicationStore;
const panes = getContext('rill:app:panes');

</script>
        
  <div class='drawer-container flex fixed'  style:top="0px" style:width="{$panes.right}px">    
    <div class='drawer-handler w-4 absolute hover:cursor-col-resize -translate-x-2 body-height'
    use:drag={{ minSize: 400, side: 'right' }} />
  
    <div class='inspector'  style="width: 100%;">
      {#if $store?.activeEntity?.type === EntityType.Model}
        <ModelInspector />
      {/if}
    </div>
  </div>
  <style lang="postcss">
  
  .inspector {
    font-size: 12px;
  }
  
  </style>