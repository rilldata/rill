<script lang="ts">
  import { onMount } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";

  export let active;
  const initialState = active;
  let mounted = false;
  let duration = 0;
  let timer;
  let element;

  let thingWidth = tweened(0, { duration: 50 });

  onMount(() => {
    const obs = new ResizeObserver(() => {
      const bbox = element.getBoundingClientRect();
      thingWidth.set(bbox.width);
    });
    obs.observe(element);
    console.log(obs, element);
    thingWidth.set(element.getBoundingClientRect().width, { duration: 0 });
    mounted = true;
  });

  function setDuration() {
    duration = 150;
    clearTimeout(timer);
    timer = setTimeout(() => {
      duration = 0;
    }, 150);
  }
  $: if ((active || !active) && mounted) setDuration();

  /**
   * because transitions in keyed blocks don't reactively update their parameters, ya gotta do it the old fashioned way
   * by hacking the css-tweening function at the moment of tween.
   */
  function hackedLocalKeyFly(
    node: Element,
    { delay = 0, easing = cubicOut, x = 0, y = 0, opacity = 0 }
  ) {
    if (!mounted) return;
    setDuration();
    const style = getComputedStyle(node);
    const target_opacity = +style.opacity;
    const transform = style.transform === "none" ? "" : style.transform;

    const od = target_opacity * (1 - opacity);

    return {
      delay,
      duration,
      easing,
      css: (t, u) => {
        return `
			transform: ${transform} translate(${
          (1 - t) * x * (duration === 0 ? 0 : 1)
        }px, ${(1 - t) * y * (duration === 0 ? 0 : 1)}px);
			opacity: ${duration === 0 ? 1 : target_opacity - od * u}`;
      },
    };
  }
</script>

<div class="relative flex gap-x-1">
  <div class="invisible absolute" bind:this={element}>
    {active ? "Hide" : "Show"}
  </div>
  <div class="invisible" style:width="{$thingWidth}px" />
  {#key active}
    <div
      class="absolute"
      style:left="0"
      style:top="0px"
      transition:hackedLocalKeyFly={{ y: 7.5 * (!active ? 1 : -1) }}
    >
      {active ? "Hide" : "Show"}
    </div>
  {/key}
  <slot />
</div>
