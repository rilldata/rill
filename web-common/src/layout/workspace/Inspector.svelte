<script lang="ts">
  import { slide } from "svelte/transition";
  import Resizer from "../Resizer.svelte";
  import { workspaces } from "./workspace-stores";

  export let filePath: string;
  export let resizable = true;
  export let fixedWidth: number | undefined = undefined;
  export let minWidth = 320;

  let resizing = false;

  $: workspace = workspaces.get(filePath);
  $: widthStore = workspace.inspector.width;
  $: visible = workspace.inspector.visible;

  $: width = fixedWidth ?? $widthStore;
</script>

{#if $visible}
  <aside
    class="inspector-wrapper"
    style:width="{width + 8}px"
    transition:slide={{ axis: "x", duration: 500 }}
    aria-label="Inspector Panel"
  >
    <Resizer
      disabled={!resizable}
      absolute={false}
      direction="EW"
      side="left"
      min={fixedWidth ?? minWidth}
      max={fixedWidth ?? 420}
      dimension={fixedWidth ?? width}
      onUpdate={(newWidth) => {
        widthStore.set(newWidth);
      }}
      bind:resizing
    />

    <div class="inner" style:width="{width}px">
      <slot />
    </div>
  </aside>
{/if}

<style lang="postcss">
  .inspector-wrapper {
    will-change: width;
    @apply h-full flex-none flex relative;
  }

  .inner {
    will-change: width;
    @apply h-full flex-none;
    @apply border border-gray-200 bg-white;
    @apply overflow-y-auto overflow-x-hidden rounded-[2px];
  }
</style>
