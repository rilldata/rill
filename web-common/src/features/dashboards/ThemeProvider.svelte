<script lang="ts">
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
  class="dashboard-theme-boundary flex flex-col size-full"
  bind:this={themeBoundary}
>
  <style bind:this={styleEl}></style>
  <slot />
</div>
