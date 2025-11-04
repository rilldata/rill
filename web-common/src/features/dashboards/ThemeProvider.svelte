<script lang="ts">
  import { Theme } from "../themes/theme";

  export let theme: Theme | undefined;

  let themeBoundary: HTMLElement | null = null;
  let styleEl: HTMLStyleElement | null = null;

  $: css = theme?.css;

  $: if (themeBoundary && styleEl && css) {
    // @ts-expect-error
    styleEl.textContent = css;
  }
</script>

<div
  class="dashboard-theme-boundary flex flex-col size-full"
  bind:this={themeBoundary}
>
  <style bind:this={styleEl}></style>
  <slot />
</div>
