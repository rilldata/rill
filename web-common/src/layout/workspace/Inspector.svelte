<script lang="ts">
  import Resizer from "../Resizer.svelte";
  import { workspaces } from "./workspace-stores";
  import { page } from "$app/stores";
  import { slide } from "svelte/transition";

  let resizing = false;

  $: context = $page.url.pathname;
  $: workspace = workspaces.get(context);
  $: width = workspace.inspector.width;
  $: visible = workspace.inspector.visible;
</script>

{#if $visible}
  <aside
    class="inspector-wrapper"
    style:width="{$width}px"
    transition:slide={{ axis: "x", duration: 500 }}
  >
    <Resizer
      direction="EW"
      side="left"
      min={300}
      max={500}
      dimension={$width}
      onUpdate={(newWidth) => {
        width.set(newWidth);
      }}
      bind:resizing
    />

    <div class="inner" style:width="{$width}px">
      <slot />
    </div>
  </aside>
{/if}

<style lang="postcss">
  .inspector-wrapper {
    will-change: width;
    @apply h-full flex-none relative;
    @apply border border-gray-200 bg-white;
    @apply overflow-y-auto overflow-x-hidden rounded-[2px];
  }

  .inner {
    will-change: width;
    @apply h-fit;
  }
</style>
