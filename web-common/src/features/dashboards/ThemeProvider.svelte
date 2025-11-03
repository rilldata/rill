<script lang="ts">
  import { page } from "$app/stores";
  import { useTheme } from "@rilldata/web-common/features/themes/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import type { V1ThemeSpec } from "@rilldata/web-common/runtime-client";
  import { Theme } from "../themes/theme";
  import { setContext } from "svelte";

  export let theme: string | V1ThemeSpec | undefined;
  export let scope: string = "explore";

  const resolvedTheme = new Theme("canvas");

  const { css } = resolvedTheme;

  setContext("themeContext", resolvedTheme);

  let themeBoundary: HTMLElement | null = null;
  let styleEl: HTMLStyleElement | null = null;
  let themeSpec: V1ThemeSpec | undefined;

  $: themeFromUrl = $page.url.searchParams.get("theme");

  $: ({ instanceId } = $runtime);

  $: themeName =
    themeFromUrl || (typeof theme === "string" ? theme : undefined);

  $: themeQuery = themeName ? useTheme(instanceId, themeName) : undefined;

  $: themeSpec =
    typeof theme !== "string" && theme ? theme : $themeQuery?.data?.theme?.spec;

  $: if (themeSpec) resolvedTheme.updateThemeSpec(themeSpec);

  $: if (themeBoundary && styleEl && $css) {
    styleEl.textContent = $css;
  }
</script>

<div
  data-theme-scope={scope}
  class="dashboard-theme-boundary"
  bind:this={themeBoundary}
>
  <style bind:this={styleEl}></style>
  <slot />
</div>

<style>
  .dashboard-theme-boundary {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }
</style>
