<script context="module" lang="ts">
  export const width = writable(0);
  export const position = writable(0);
</script>

<script lang="ts">
  import ProjectGlobalStatusIndicator from "@rilldata/web-admin/features/projects/status/ProjectGlobalStatusIndicator.svelte";
  import { onMount } from "svelte";
  import { writable } from "svelte/store";

  export let route: string;
  export let label: string;
  export let selected: boolean;
  export let organization: string;
  export let project: string | undefined = undefined;

  let size: number = 0;
  let left: number = 0;

  $: if (selected && size) width.set(size);
  $: if (selected && left) position.set(left);

  const observer = new ResizeObserver((entries) => {
    for (const entry of entries) {
      //   size = entry.contentRect.width;
      if (entry.target instanceof HTMLElement) {
        //   size = entry.clientWidth;
        left = entry.target.offsetLeft;
        size = entry.target.clientWidth;
      }
    }
  });

  let element: HTMLAnchorElement;

  onMount(() => {
    observer.observe(element);
  });
</script>

<a href={route} class:selected bind:this={element}>
  <p data-content={label}>
    {label}
  </p>
  {#if label === "Status" && project}
    <ProjectGlobalStatusIndicator {organization} {project} />
  {/if}
</a>

<style lang="postcss">
  a {
    @apply px-2 py-1.5 flex gap-x-1 items-center w-fit;
    @apply rounded-sm text-gray-500;
    @apply text-xs font-medium justify-center;
  }

  .selected {
    @apply text-gray-900 font-semibold;
  }

  a:hover {
    @apply bg-slate-100 text-gray-700;
  }

  p {
    @apply text-center;
  }

  /* Prevent layout shift on font weight change */
  p::before {
    content: attr(data-content);
    display: block;
    font-weight: 600;
    height: 0;
    visibility: hidden;
  }
</style>
