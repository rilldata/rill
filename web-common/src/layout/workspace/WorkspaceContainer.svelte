<script lang="ts">
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { SIDE_PAD } from "../config";
  import Inspector from "./Inspector.svelte";

  export let inspector = true;
  export let bgClass = "bg-gray-100";

  const navigationWidth = getContext<Writable<number>>(
    "rill:app:navigation-width-tween",
  );
  const navVisibilityTween = getContext<Writable<number>>(
    "rill:app:navigation-visibility-tween",
  );
</script>

<div
  class="flex flex-col h-screen overflow-hidden absolute"
  style:left="{($navigationWidth || 0) * (1 - $navVisibilityTween)}px"
  style:padding-left="{$navVisibilityTween * SIDE_PAD}px"
  style:right="0px"
>
  {#if $$slots.header}
    <header class="bg-white w-full h-fit">
      <slot name="header" />
    </header>
  {/if}

  <div class="h-full {bgClass} w-full flex overflow-hidden">
    <div class="w-full h-full overflow-hidden">
      <slot name="body" />
    </div>
    {#if inspector}
      <Inspector>
        <slot name="inspector" />
      </Inspector>
    {/if}
  </div>
</div>
