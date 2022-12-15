<script lang="ts">
  import { cubicOut } from "svelte/easing";

  export let active;
  let mounted = false;
  let duration = 0;
  let timer;

  function setDuration() {
    if (!mounted) {
      mounted = true;
      return;
    }
    duration = 150;
    clearTimeout(timer);
    timer = setTimeout(() => {
      duration = 0;
    }, 150);
  }

  $: setDuration();

  /**
   * because transitions in keyed blocks don't reactively update their parameters, ya gotta do it the old fashioned way
   * by hacking the css-tweening function at the moment of tween.
   */
  function hackedLocalKeyFly(
    node: Element,
    { delay = 0, easing = cubicOut, x = 0, y = 0, opacity = 0 }
  ) {
    const style = getComputedStyle(node);
    const target_opacity = +style.opacity;
    const transform = style.transform === "none" ? "" : style.transform;

    const od = target_opacity * (1 - opacity);

    return {
      delay,
      duration,
      easing,
      css: (t, u) => `
			transform: ${transform} translate(${
        (1 - t) * x * (duration === 0 ? 0 : 1)
      }px, ${(1 - t) * y * (duration === 0 ? 0 : 1)}px);
			opacity: ${duration === 0 ? 1 : target_opacity - od * u}`,
    };
  }
</script>

<div class="relative">
  <span class="invisible">{active ? "Hide" : "Show"}</span>
  {#key active}
    <span
      class="absolute"
      style:left="0"
      style:top="0px"
      transition:hackedLocalKeyFly={{ y: 7.5 * (!active ? 1 : -1) }}
      >{active ? "Hide" : "Show"}</span
    >
  {/key}
  <slot />
</div>
