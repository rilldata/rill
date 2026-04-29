<script context="module" lang="ts">
  export const width = writable(0);
  export const position = writable(0);
</script>

<script lang="ts">
  import ProjectGlobalStatusIndicator from "@rilldata/web-admin/features/projects/status/overview/ProjectGlobalStatusIndicator.svelte";
  import { onMount } from "svelte";
  import { writable } from "svelte/store";

  export let route: string;
  export let label: string;
  export let selected: boolean;
  export let organization: string;
  export let project: string | undefined = undefined;

  let size: number = 0;
  let left: number = 0;
  let element: HTMLAnchorElement;

  $: if (selected && size) {
    width.set(size);
    position.set(left);
  }

  function measure() {
    if (!element) return;
    left = element.offsetLeft;
    size = element.clientWidth;
  }

  onMount(() => {
    // Observe self for size changes AND the parent nav so we re-measure when
    // sibling tabs mount/unmount (feature flags, permissions). ResizeObserver
    // does not fire on pure offsetLeft changes, so without this the underline
    // sticks to a stale position after sibling layout shifts.
    const observer = new ResizeObserver(measure);
    observer.observe(element);
    if (element.parentElement) observer.observe(element.parentElement);
    return () => observer.disconnect();
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
    @apply rounded-sm text-fg-muted;
    @apply text-xs font-medium justify-center;
  }

  .selected {
    @apply text-fg-accent font-semibold;
  }

  a:hover {
    @apply bg-gray-100;
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
