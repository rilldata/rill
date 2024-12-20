<script context="module" lang="ts">
  import { goto } from "$app/navigation";
  import * as ContextMenu from "@rilldata/web-common/components/context-menu";
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
</script>

<script lang="ts">
  const dispatch = createEventDispatcher();

  export let i: number;
  export let gapSize: number;
  export let component: V1CanvasItem;
  export let selected: boolean;
  export let interacting: boolean;
  export let width: number;
  export let height: number;
  export let top: number;
  export let left: number;
  export let radius: number;
  export let scale: number;
  export let instanceId: string;

  $: componentName = component?.component;
  $: inlineComponent = component?.definedInCanvas;

  $: finalLeft = width < 0 ? left + width : left;
  $: finalTop = height < 0 ? top + height : top;
  $: finalWidth = Math.abs(width);
  $: finalHeight = Math.abs(height);
  $: padding = gapSize;

  function handlePointerOver(e: PointerEvent) {
    dispatch("pointerover", { index: i });
  }

  function handlePointerOut(e: PointerEvent) {
    dispatch("pointerout", { index: null });
  }

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    dispatch("change", {
      e,
      dimensions: [width, height],
      position: [finalLeft, finalTop],
      changeDimensions: [0, 0],
      changePosition: [1, 1],
    });
  }
</script>

{#if componentName && !inlineComponent}
  <ContextMenu.Root>
    <ContextMenu.Trigger asChild let:builder>
      <CanvasComponent
        {instanceId}
        {i}
        {interacting}
        {componentName}
        {padding}
        {radius}
        {scale}
        {selected}
        builders={[builder]}
        height={finalHeight}
        left={finalLeft}
        top={finalTop}
        width={finalWidth}
        on:change
        on:contextmenu
        on:mousedown={handleMouseDown}
        on:pointerover={handlePointerOver}
        on:pointerout={handlePointerOut}
      />
    </ContextMenu.Trigger>

    <ContextMenu.Content class="z-[100]">
      <ContextMenu.Item
        on:click={async () => {
          await goto(`/files/charts/${componentName}.yaml`);
        }}
      >
        Go to {componentName}.yaml
      </ContextMenu.Item>
      <ContextMenu.Item on:click={() => dispatch("delete", { index: i })}
        >Delete from dashboard</ContextMenu.Item
      >
    </ContextMenu.Content>
  </ContextMenu.Root>
{:else if componentName}
  <ContextMenu.Root>
    <ContextMenu.Trigger asChild let:builder>
      <CanvasComponent
        {instanceId}
        {i}
        {interacting}
        {componentName}
        {padding}
        {radius}
        {scale}
        {selected}
        builders={[builder]}
        height={finalHeight}
        left={finalLeft}
        top={finalTop}
        width={finalWidth}
        on:change
        on:contextmenu
        on:mousedown={handleMouseDown}
        on:pointerover={handlePointerOver}
        on:pointerout={handlePointerOut}
      />
    </ContextMenu.Trigger>

    <ContextMenu.Content class="z-[100]">
      <ContextMenu.Item on:click={() => dispatch("delete", { index: i })}
        >Delete from dashboard</ContextMenu.Item
      >
    </ContextMenu.Content>
  </ContextMenu.Root>
{/if}
