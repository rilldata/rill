<script lang="ts">
  import { cubicOut as easing } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import type { BarAndLabelTweenParameterOptions } from "./types";

  export let value = 0;
  export let color;
  export let showBackground = true;
  export let compact = false;
  export let showHover = false;
  export let customBackgroundColor = undefined;
  export let justify: string | boolean = "end"; // or left
  export let tweenParameters: BarAndLabelTweenParameterOptions<number> = {
    duration: 500,
    easing,
  };

  let finalParameters: BarAndLabelTweenParameterOptions<number> = {
    ...{ duration: 500, easing },
    ...tweenParameters,
  };

  const valueTween = tweened(0, finalParameters);
  $: valueTween.set(value);
  /** for the tailwind compiler: we're creating these optional classes */
  // justify-items-stretch justify-items-end justify-items-start
  // justify-stretch justify-end -justify-start pl-2 pr-2 pl-1 pr-1
</script>

<div
  class="
    text-right grid items-center 
    {justify ? `justify-${justify}` : ''} 
    {justify ? `justify-items-${justify}` : ''} relative w-full
    {showHover ? 'hover:bg-gray-100 hover:dark:bg-gray-600' : undefined}
    {customBackgroundColor
    ? customBackgroundColor
    : showBackground
    ? 'bg-gray-100 dark:bg-gray-700'
    : 'bg-transparent'}
    "
  style:flex="1"
>
  <div
    class:pl-2={!compact}
    class:pr-2={!compact}
    class:pr-1={compact}
    class:pl-1={compact}
    class="text-right overflow-x-hidden "
    style="position: relative;"
  >
    <slot />
  </div>
  <div
    class="number-bar {color} mix-blend-multiply dark:mix-blend-screen"
    style="--width: {Math.min(1, $valueTween)};"
  />
</div>

<style lang="postcss">
  .number-bar {
    --width: 0%;
    content: "";
    display: inline-block;
    width: calc(100% * var(--width));
    position: absolute;
    left: 0;
    top: 0;
    height: 100%;
    pointer-events: none;
  }
</style>
