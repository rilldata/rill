<script lang="ts" context="module">
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import Toolbar from "./Toolbar.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import { hideBorder } from "./layout-util";
  import { onMount, onDestroy } from "svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import {
    COMPONENT_PATH_ROW_INDEX,
    COMPONENT_PATH_COLUMN_INDEX,
  } from "./stores/canvas-entity";
  import {
    getComponentThemeOverrides,
    generateComponentThemeCSS,
  } from "./utils/component-colors";
</script>

<script lang="ts">
  export let component: BaseCanvasComponent;

  let observer: IntersectionObserver | null = null;

  onMount(() => {
    if (!component) return;

    observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && component) {
          component.visible.set(true);
          if (observer) {
            observer.unobserve(container);
          }
        }
      },
      {
        root: document.querySelector(".dashboard-theme-boundary"),
        rootMargin: "120px",
        threshold: 0,
      },
    );
    if (container) {
      observer.observe(container);
    }
  });

  onDestroy(() => {
    if (observer) {
      observer.disconnect();
      observer = null;
    }
    if (styleElement) {
      styleElement.remove();
      styleElement = null;
    }
  });
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
  let styleElement: HTMLStyleElement | null = null;

  // Extract component name safely
  let componentName = "";
  let renderer = "";

  $: {
    if (component) {
      componentName = component.id || "";
      renderer = component.type || "";
    }
  }

  $: allowBorder = !hideBorder.has(renderer);

  // Get component position from pathInYAML with bounds checking
  $: rowIndex = (() => {
    if (!component) return -1;
    const idx = component.pathInYAML?.[COMPONENT_PATH_ROW_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();

  $: columnIndex = (() => {
    if (!component) return -1;
    const idx = component.pathInYAML?.[COMPONENT_PATH_COLUMN_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();

  // Get canvas spec to access item properties with null checks
  // Extract store reference first, then subscribe to it
  $: specStore = component?.parent?.specStore;
  $: canvasData = specStore ? $specStore?.data : undefined;
  $: canvasRows = canvasData?.canvas?.rows ?? [];

  // Get the item for this component with bounds checking
  $: item = (() => {
    if (
      rowIndex < 0 ||
      columnIndex < 0 ||
      !canvasRows ||
      rowIndex >= canvasRows.length
    ) {
      return undefined;
    }
    const row = canvasRows[rowIndex];
    if (!row?.items || columnIndex >= row.items.length) {
      return undefined;
    }
    return row.items[columnIndex];
  })();

  // Detect dark mode (SSR-safe)
  $: isDarkMode =
    typeof window !== "undefined" ? $themeControl === "dark" : false;

  // Get component theme overrides
  $: themeOverrides = getComponentThemeOverrides(item);

  // Generate scoped CSS for component theme overrides
  $: themeCSS = componentName
    ? generateComponentThemeCSS(componentName, themeOverrides)
    : "";

  // Function to update style element
  function updateStyleElement() {
    if (typeof window === "undefined" || !componentName) {
      return;
    }

    if (themeCSS) {
      if (!styleElement) {
        styleElement = document.createElement("style");
        styleElement.setAttribute("data-component-theme", componentName);
        document.head.appendChild(styleElement);
      }
      styleElement.textContent = themeCSS;
    } else if (styleElement) {
      styleElement.remove();
      styleElement = null;
    }
  }

  // Update style element when themeCSS or componentName changes
  $: if (componentName) {
    updateStyleElement();
  }
</script>

<article
  bind:this={container}
  role="presentation"
  id={componentName}
  class:selected
  class:editable
  class:opacity-20={ghost}
  style:pointer-events={!allowPointerEvents ? "none" : "auto"}
  style:background-color="var(--card)"
  class:outline={allowBorder || open}
  class:shadow-sm={allowBorder || open}
  class="group component-card size-full flex flex-col cursor-pointer z-10 p-0 relative outline-[1px] overflow-hidden rounded-sm"
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
    on:mousedown={onMouseDown}
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

  .component-card:has(.component-error) {
    @apply outline-red-200;
  }

  .selected {
    @apply shadow-md outline-primary-400 outline-[1.5px];

    outline-style: solid !important;
  }
</style>
