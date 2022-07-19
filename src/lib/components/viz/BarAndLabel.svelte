<script lang="ts">
  import { tweened } from "svelte/motion";
  import { cubicOut as easing } from "svelte/easing";
  import type { BarAndLabelTweenParameterOptions } from "./types";
  export let value = 0;
  export let color;
  export let showBackground = true;
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
  // justify-stretch justify-end -justify-start
</script>

<div
  class="
    text-right grid items-center 
    {justify ? `justify-${justify}` : ''} 
    {justify ? `justify-items-${justify}` : ''} relative w-full"
  style:background-color={showBackground
    ? "hsla(217,5%, 90%, .25)"
    : "hsl(217, 0%, 100%, .25)"}
>
  <div class="pl-2 pr-2 text-right" style="position: relative;"><slot /></div>
  <div
    class="number-bar {color}"
    style="--width: {Math.min(1, $valueTween)};"
  />
</div>

<style>
  .number-bar {
    --width: 0%;
    content: "";
    display: inline-block;
    width: calc(100% * var(--width));
    position: absolute;
    left: 0;
    top: 0;
    height: 100%;

    mix-blend-mode: multiply;
    pointer-events: none;
  }
</style>
