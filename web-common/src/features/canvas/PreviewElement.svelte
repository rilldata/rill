<script context="module" lang="ts">
  import { goto } from "$app/navigation";
  import * as ContextMenu from "@rilldata/web-common/components/context-menu";
  import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import CanvasComponent from "./CanvasComponent.svelte";
</script>

<script lang="ts">
  export let i: number;
  export let component: V1CanvasItem;
  export let selected: boolean;
  export let interacting: boolean;
  export let width: number;
  export let height: number;
  export let top: number;
  export let left: number;
  export let instanceId: string;

  $: componentName = component?.component;
  $: inlineComponent = component?.definedInCanvas;

  const dispatch = createEventDispatcher();

  $: finalLeft = width < 0 ? left + width : left;
  $: finalTop = height < 0 ? top + height : top;
  $: finalWidth = Math.abs(width);
  $: finalHeight = Math.abs(height);

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    dispatch("mousedown", {
      e,
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
        {selected}
        builders={[builder]}
        height={finalHeight}
        left={finalLeft}
        top={finalTop}
        width={finalWidth}
        on:change
        on:contextmenu
        on:mousedown={handleMouseDown}
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
        {selected}
        builders={[builder]}
        height={finalHeight}
        left={finalLeft}
        top={finalTop}
        width={finalWidth}
        on:change
        on:contextmenu
        on:mousedown={handleMouseDown}
      />
    </ContextMenu.Trigger>

    <ContextMenu.Content class="z-[100]">
      <ContextMenu.Item on:click={() => dispatch("delete", { index: i })}
        >Delete from dashboard</ContextMenu.Item
      >
    </ContextMenu.Content>
  </ContextMenu.Root>
{/if}
