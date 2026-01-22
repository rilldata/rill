<script lang="ts">
  import { dynamicHeight } from "@rilldata/web-common/layout/layout-settings.ts";
  import { Theme } from "../themes/theme";

  export let theme: Theme | undefined;

  let themeBoundary: HTMLElement | null = null;
  let styleEl: HTMLStyleElement | null = null;

  $: css = theme?.css;

  // Update theme CSS, or clear it when theme is undefined
  $: if (themeBoundary && styleEl) {
    // @ts-expect-error - textContent is writable but typed as readonly in some environments
    styleEl.textContent = css || "";
  }
</script>

<div
  class="dashboard-theme-boundary flex flex-col overflow-hidden"
  bind:this={themeBoundary}
  class:w-full={$dynamicHeight}
  class:size-full={!$dynamicHeight}
>
  <style bind:this={styleEl}></style>
  <slot />
</div>
