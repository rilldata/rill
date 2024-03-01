<script context="module" lang="ts">
  export type Shortcut = {
    description: string;
    keys?: ShortcutYeah["$$prop_def"]["keys"];
    modifiers?: ShortcutYeah["$$prop_def"]["modifiers"];
    click?: boolean;
    condition?: boolean;
  };
</script>

<script lang="ts">
  import TooltipTitle from "./TooltipTitle.svelte";
  import ShortcutYeah from "./ShortcutYeah.svelte";

  export let title: string;
  export let text: string;
  export let shortcuts: Shortcut[] = [];
  export let x: number;
  export let y: number;
  export let alignment: "start" | "center" | "end" = "start";
  export let position: "top" | "bottom" | "left" | "right" = "top";
  export let maxWidth = "400px";
</script>

<div
  class="{position} {alignment} tooltip"
  style:left="{x}px"
  style:top="{y}px"
  style:max-width={maxWidth}
>
  {#if title}
    <TooltipTitle>
      <svelte:fragment slot="name">{title}</svelte:fragment>
      <svelte:fragment slot="description">{text}</svelte:fragment>
    </TooltipTitle>
  {/if}

  {#each shortcuts as shortcut (shortcut)}
    {#if shortcut.condition === undefined || shortcut.condition}
      <div class="flex justify-between gap-x-2 max-w-full">
        <p>{shortcut.description}</p>
        <ShortcutYeah {...shortcut} />
      </div>
    {/if}
  {/each}
</div>

<style lang="postcss">
  .tooltip {
    @apply absolute pointer-events-none;
    @apply bg-gray-700 text-white rounded;
    @apply px-2 py-1;
    @apply flex flex-col gap-1;
    @apply w-fit overflow-hidden;
  }

  .top {
    @apply -translate-y-full;
  }

  .left {
    @apply -translate-x-full;
  }

  .center.left,
  .center.right {
    @apply -translate-y-1/2;
  }

  .center.top,
  .center.bottom {
    @apply -translate-x-1/2;
  }
</style>
