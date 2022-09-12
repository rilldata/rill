<script lang="ts">
  import type { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";
  import CollapsibleTableHeader from "./CollapsibleTableHeader.svelte";

  export let entityType: EntityType;
  export let name: string;
  export let cardinality: number = undefined;
  export let showRows = true;
  export let sizeInBytes: number = undefined;
  export let active = false;
  export let draggable = true;
  export let show = false;
  export let showTitle = true;
  export let notExpandable = false;

  let containerWidth = 0;
  let contextMenuOpen;
  let container;

  onMount(() => {
    const observer = new ResizeObserver(() => {
      containerWidth = container?.clientWidth ?? 0;
    });
    observer.observe(container);
    return () => observer.unobserve(container);
  });
</script>

<div bind:this={container}>
  {#if showTitle}
    <div {draggable} class="active:cursor-grabbing">
      <CollapsibleTableHeader
        on:select
        on:query
        on:expand={() => (show = !show)}
        bind:contextMenuOpen
        bind:name
        bind:show
        {showRows}
        {entityType}
        {cardinality}
        {sizeInBytes}
        {active}
        {notExpandable}
      >
        <slot name="header-buttons" />
        <svelte:fragment slot="menu-items" let:toggleMenu>
          <slot name="menu-items" {toggleMenu} />
        </svelte:fragment>
      </CollapsibleTableHeader>
    </div>
  {/if}

  {#if show}
    <div
      class="pt-1 pb-3 pl-accordion"
      transition:slide|local={{ duration: 120 }}
    >
      <slot name="summary" {containerWidth} />
    </div>
  {/if}
</div>
