<script lang="ts" context="module">
  export const THEME_STORE_CONTEXT_KEY = Symbol("theme-store");
</script>

<script lang="ts">
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import { Theme } from "../themes/theme";

  export let theme: Theme | undefined;
  /**
   * Whether to apply full-width/height sizing classes.
   * Set to false when the parent already controls dimensions (e.g., sidebar chat).
   */
  export let applyLayout: boolean = true;

  let themeBoundary: HTMLElement | null = null;
  let styleEl: HTMLStyleElement | null = null;

  const themeStore = writable<Theme | undefined>(theme);
  $: themeStore.set(theme);

  setContext(THEME_STORE_CONTEXT_KEY, themeStore);

  $: css = theme?.css;

  // Update theme CSS, or clear it when theme is undefined
  $: if (themeBoundary && styleEl) {
    // @ts-expect-error - textContent is writable but typed as readonly in some environments
    styleEl.textContent = css || "";
  }
</script>

<div
  class="dashboard-theme-boundary flex flex-col overflow-hidden"
  class:w-full={applyLayout && $dynamicHeight}
  class:size-full={applyLayout && !$dynamicHeight}
  bind:this={themeBoundary}
>
  <style bind:this={styleEl}></style>
  <slot />
</div>
