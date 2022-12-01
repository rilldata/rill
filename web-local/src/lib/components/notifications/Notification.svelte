<script>
  import { scale } from "svelte/transition";

  export let width = undefined;
  export let level = "info"; // info, error.

  export let xOffset = undefined;
  export let yOffset = undefined;

  // positioning elements.
  export let location;

  // if specified, set the width variable.

  function makeStyle({ width, xOffset, yOffset }) {
    let styles = [
      width ? `--width: ${width}` : undefined,
      xOffset ? `--x-offset: ${xOffset}` : undefined,
      yOffset ? `--y-offset: ${yOffset}` : undefined,
    ];
    return styles.filter((d) => d !== undefined).join("; ");
  }

  $: style = makeStyle({ width, xOffset, yOffset });
</script>

<aside
  transition:scale={{ duration: 200, start: 0.98, opacity: 0 }}
  {style}
  class="whitespace-pre radius-sm notification-{level} 
      {location !== undefined
    ? `notification-floating notification-floating-${location}`
    : ''}"
>
  <div class="icon centered">
    <slot name="icon" />
  </div>
  <div class="body centered">
    <slot name="body" />
    <slot />
  </div>
  <div class="cta centered">
    <slot name="cta" />
  </div>
</aside>

<style lang="postcss">
  aside {
    --width: max-content;
    padding: 8px 14px;
    width: var(--width);
    box-shadow: var(--rally-box-shadow-sm);
    font-style: normal;
    /* font-weight: 600; */
    font-size: 15px;
    line-height: 22px;
    display: grid;
    grid-column-gap: 10px;
    grid-template-columns: max-content auto max-content;
    align-items: center;
    align-content: center;
    @apply shadow-lg rounded-md;
  }

  .notification-info,
  .notification-info > * {
    color: #0c0c0d;
    @apply bg-slate-800 text-slate-100;
  }

  .notification-error,
  .notification-error > * {
    color: var(--color-white);
    background-color: var(--color-red-60);
  }

  .icon {
    font-size: 24px;
    height: 24px;
  }

  .centered {
    height: max-content;
    display: grid;
    align-self: center;
    align-items: center;
    align-content: center;
  }

  .centered > * {
    height: max-content;
    display: grid;
    align-self: center;
    align-items: center;
    align-content: center;
  }

  .notification-floating {
    position: fixed;
    --pad: 20px;
    /* stylelint-disable */
    --x-offset: 0px;
    --y-offset: 0px;
    /* stylelint-enable */
    --x-pad: calc(var(--pad) + var(--x-offset));
    --y-pad: calc(var(--pad) + var(--y-offset));
  }

  .notification-floating-bottom,
  .notification-floating-bottom-center {
    bottom: var(--y-pad);
    left: calc(50% + var(--x-offset));
    transform: translateX(-50%);
  }

  .notification-floating-bottom-right {
    right: var(--x-pad);
    bottom: var(--y-pad);
  }

  .notification-floating-bottom-left {
    left: var(--x-pad);
    bottom: var(--y-pad);
  }

  .notification-floating-top,
  .notification-floating-top-center {
    top: var(--y-pad);
    left: calc(50% + var(--x-offset));
    transform: translateX(-50%);
  }

  .notification-floating-top-left {
    left: var(--x-pad);
    top: var(--y-pad);
  }

  .notification-floating-top-right {
    right: var(--x-pad);
    top: var(--y-pad);
  }
</style>
