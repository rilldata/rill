<script lang="ts">
import { getContext } from "svelte";
import ModelInspector from "./Model.svelte";

import type { ApplicationStore } from "$lib/app-store";

import { drag } from "$lib/drag";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { inspectorVisibilityTween, inspectorVisible } from "$lib/layout-store";
import Portal from "$lib/components/Portal.svelte";
const store = getContext('rill:app:store') as ApplicationStore;
const layout = getContext('rill:app:layout');

</script>

  <div 
      class='
        border-l 
        border-transparent 
        fixed 
        overflow-auto 
        hover:border-gray-200 
        transition-colors
        body-height
      ' 
      class:hidden={$inspectorVisibilityTween === 1}
      class:pointer-events-none={!$inspectorVisible}
      style:top="0px"
      style:width="{$layout.inspectorWidth}px"
    >    
    {#if $inspectorVisible}
      <Portal>
        <div 
          class='fixed drawer-handler w-4 hover:cursor-col-resize translate-x-2 body-height' 
          style:right="{(1- $inspectorVisibilityTween) * $layout.inspectorWidth}px"
          use:drag={{ minSize: 400, side: 'inspectorWidth', reverse: true }} />
      </Portal>
    {/if}
    
  
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