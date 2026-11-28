<script lang="ts">
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { tryUseRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import { Theme } from "../themes/theme";
  import { THEME_STORE_CONTEXT_KEY } from "../themes/theme-boundary";
  import {
    loadThemeFont,
    THEME_CHART_FONT_CONTEXT_KEY,
    type ResolvedThemeFont,
  } from "../themes/theme-font";
  import { themeManager } from "../themes/theme-manager";

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

  const client = tryUseRuntimeClient();

  // The theme's font for charts. HTML text picks the font up from the theme's CSS as soon as
  // the browser has it; charts are handed it only once its files have loaded (see getThemeChartFont).
  const chartFont = writable<string | undefined>(undefined);
  setContext(THEME_CHART_FONT_CONTEXT_KEY, chartFont);

  let fontLoadId = 0;
  async function loadFont(font: ResolvedThemeFont | undefined) {
    const id = ++fontLoadId;
    if (font && client) await loadThemeFont(client, font);
    // A newer theme may have arrived while this one's files were downloading.
    if (id === fontLoadId) chartFont.set(font?.cssFamily);
  }

  $: void loadFont(theme?.font);

  $: css = theme?.css;

  // Update theme CSS, or clear it when theme is undefined
  $: if (themeBoundary && styleEl) {
    // @ts-expect-error - textContent is writable but typed as readonly in some environments
    styleEl.textContent = css || "";
    // Clear cached CSS variable lookups so charts re-resolve against the new theme
    themeManager.clearCSSVariableCache();
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
