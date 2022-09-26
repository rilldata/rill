<script lang="ts">
  import { EntityType } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
  import type { ApplicationStore } from "../../application-state-stores/application-store";
  import {
    inspectorVisibilityTween,
    inspectorVisible,
    layout,
  } from "../../application-state-stores/layout-store";
  import Portal from "../Portal.svelte";
  import { drag } from "../../drag";
  import { getContext } from "svelte";
  import ModelInspector from "./model/ModelInspector.svelte";
  import SourceInspector from "./SourceInspector.svelte";

  const store = getContext("rill:app:store") as ApplicationStore;
</script>

<div
  class="
      bg-white
        border-l 
        border-transparent 
        fixed 
        overflow-auto 
        hover:border-gray-200 
        transition-colors
        h-screen
      "
  class:hidden={$inspectorVisibilityTween === 1}
  class:pointer-events-none={!$inspectorVisible}
  style:top="0px"
  style:width="{$layout.inspectorWidth}px"
>
  <!-- draw handler -->
  {#if $inspectorVisible}
    <Portal>
      <div
        class="fixed drawer-handler w-4 hover:cursor-col-resize translate-x-2 h-screen"
        style:right="{(1 - $inspectorVisibilityTween) *
          $layout.inspectorWidth}px"
        use:drag={{ minSize: 300, side: "inspectorWidth", reverse: true }}
      />
    </Portal>
  {/if}

  <div style="width: 100%;">
    {#if $store?.activeEntity?.type === EntityType.Model}
      <!-- re-render if the id changes. -->
      {#key $store?.activeEntity?.id}
        <ModelInspector />
      {/key}
    {:else if $store?.activeEntity?.type === EntityType.Table}
      {#key $store?.activeEntity?.id}
        <SourceInspector />
      {/key}
    {/if}
  </div>
</div>
