<script lang="ts" context="module">
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import Toolbar from "./Toolbar.svelte";
  import type { BaseCanvasComponent } from "./components/BaseCanvasComponent";
  import { hideBorder } from "./layout-util";
  import { onMount } from "svelte";
  import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
  import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";
  import {
    COMPONENT_PATH_ROW_INDEX,
    COMPONENT_PATH_COLUMN_INDEX,
  } from "./stores/canvas-entity";
</script>

<script lang="ts">
  const observer = new IntersectionObserver(
    ([entry]) => {
      if (entry.isIntersecting) {
        component.visible.set(true);
        observer.unobserve(container);
      }
    },
    {
      root: document.querySelector(".dashboard-theme-boundary"),
      rootMargin: "120px",
      threshold: 0,
    },
  );

  onMount(() => {
    observer.observe(container);
  });

  export let component: BaseCanvasComponent;
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

  // Get component position from pathInYAML with bounds checking
  $: rowIndex = (() => {
    const idx = component.pathInYAML?.[COMPONENT_PATH_ROW_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();

  $: columnIndex = (() => {
    const idx = component.pathInYAML?.[COMPONENT_PATH_COLUMN_INDEX];
    return typeof idx === "number" && idx >= 0 ? idx : -1;
  })();

  // Get canvas spec to access item properties with null checks
  // Extract store reference first, then subscribe to it
  $: specStore = component.parent?.specStore;
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

  // Get background colors from item spec (camelCase from generated types)
  $: backgroundColorLight = item?.backgroundColorLight?.trim() || undefined;
  $: backgroundColorDark = item?.backgroundColorDark?.trim() || undefined;

  // Detect dark mode (SSR-safe)
  $: isDarkMode =
    typeof window !== "undefined" ? $themeControl === "dark" : false;

  // Resolve background color: use override if set, otherwise use theme's card color
  $: backgroundColor = (() => {
    // Use override if available for current mode
    if (isDarkMode && backgroundColorDark) {
      return backgroundColorDark;
    }
    if (!isDarkMode && backgroundColorLight) {
      return backgroundColorLight;
    }

    // Fallback to theme's card color with proper fallback
    if (typeof window !== "undefined") {
      const cardColor = themeManager.resolveCSSVariable(
        "var(--card)",
        isDarkMode,
      );
      // If resolved color is still a CSS variable, use a safe default
      if (cardColor && !cardColor.startsWith("var(")) {
        return cardColor;
      }
    }

    // Ultimate fallback (should rarely be needed)
    return undefined;
  })();
</script>

<article
  bind:this={container}
  role="presentation"
  id={componentName}
  class:selected
  class:editable
  class:opacity-20={ghost}
  style:pointer-events={!allowPointerEvents ? "none" : "auto"}
  style:background-color={backgroundColor}
  class:outline={allowBorder || open}
  class:shadow-sm={allowBorder || open}
  class="group component-card size-full flex flex-col cursor-pointer z-10 p-0 relative outline-[1px] outline-gray-200 dark:outline-gray-300 overflow-hidden rounded-sm"
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
