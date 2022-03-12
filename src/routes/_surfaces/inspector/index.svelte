<script lang="ts">
import { getContext } from "svelte";
import ModelInspector from "./Model.svelte";

import type { ApplicationStore } from "$lib/app-store";

import { drag } from "$lib/drag";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { inspectorVisibilityTween, inspectorVisible, layout } from "$lib/layout-store";
import Portal from "$lib/components/Portal.svelte";

const store = getContext('rill:app:store') as ApplicationStore;

</script>

  <div 
      class='
      bg-white
        border-l 
        border-transparent 
        fixed 
        overflow-auto 
        hover:border-gray-200 
        transition-colors
        h-screen
      ' 
      class:hidden={$inspectorVisibilityTween === 1}
      class:pointer-events-none={!$inspectorVisible}
      style:top="0px"
      style:width="{$layout.inspectorWidth}px"
    > 
    <!-- draw handler -->
    {#if $inspectorVisible}
      <Portal>
        <div 
          class='fixed drawer-handler w-4 hover:cursor-col-resize translate-x-2 h-screen' 
          style:right="{(1- $inspectorVisibilityTween) * $layout.inspectorWidth}px"
          use:drag={{ minSize: 400, side: 'inspectorWidth', reverse: true }} />
      </Portal>
    {/if}
    
  
    <div style="width: 100%;">
      {#if $store?.activeEntity?.type === EntityType.Model}
          <!-- re-render if the id changes. -->
          {#key $store?.activeEntity?.id}
            <ModelInspector />
          {/key}
      {/if}
    </div>
  </div>
 