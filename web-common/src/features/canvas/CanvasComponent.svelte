<script lang="ts" context="module">
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { onMount } from "svelte";
  import Toolbar from "./Toolbar.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import { hideBorder } from "./layout-util";
</script>

<script lang="ts">
  import { get } from "svelte/store";

  let observer: IntersectionObserver;

  let mounted = false;

  onMount(() => {
    observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          component.visible.set(true);
          observer.unobserve(container);
        }
      },
      {
        root: container.closest(".dashboard-theme-boundary"),
        rootMargin: "120px",
        threshold: 0,
      },
    );
    mounted = true;
    observer.observe(container);
  });

  export let component: BaseCanvasComponent;

  let prevComponent: BaseCanvasComponent | undefined;
  $: if (mounted && component !== prevComponent) {
    const wasVisible = prevComponent ? get(prevComponent.visible) : false;
    prevComponent = component;
    if (wasVisible) {
      component.visible.set(true);
    } else {
      observer.unobserve(container);
      observer.observe(container);
    }
  }
  export let selected = false;
  export let ghost = false;
  export let allowPointerEvents = true;
  export let editable = false;
  export let navigationEnabled: boolean = true;
  export let onMouseDown: (e: MouseEvent) => void = () => {};
  export let onDuplicate: () => void = () => {};
  export let onDelete: () => void = () => {};

  let open = false;
  let container: HTMLElement;

  $: ({ id: componentName, type: renderer } = component);

  $: allowBorder = !hideBorder.has(renderer);
</script>

<article
  bind:this={container}
  role="presentation"
  id={componentName}
  class:selected
  class:editable
  class:opacity-20={ghost}
  style:pointer-events={!allowPointerEvents ? "none" : "auto"}
  class:outline={allowBorder || open}
  class:shadow-sm={allowBorder || open}
  class="group component-card size-full flex flex-col cursor-pointer z-10 p-0 relative outline-[1px] outline-border bg-surface-card overflow-hidden rounded-sm"
>
  <Toolbar
    {component}
    {onDelete}
    {onDuplicate}
    {editable}
    bind:dropdownOpen={open}
    {navigationEnabled}
  />

  <div
    role="presentation"
    class="size-full grow flex flex-col"
    onmousedown={onMouseDown}
  >
    {#if component}
      <svelte:component this={component.component} {component} />
    {:else}
      <div class="size-full grid place-content-center">
        <LoadingSpinner size="36px" />
      </div>
    {/if}
  </div>
</article>

<style lang="postcss">
  .component-card.editable:hover {
    @apply shadow-md outline;
  }

  .component-card:has(:global(.component-error)) {
    @apply outline-destructive;
  }

  .selected {
    @apply shadow-md outline-primary-400 outline-[1.5px];

    outline-style: solid !important;
  }
</style>
