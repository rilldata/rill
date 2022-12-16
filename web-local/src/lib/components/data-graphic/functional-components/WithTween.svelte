<!--
@component
A convenience function that takes a value, and through slot props exposes a tweened version
of the value.
-->
<script lang="ts">
  import { tweened } from "svelte/motion";
  import type { EasingFunction } from "svelte/transition";
  export let value: any;
  export let tweenProps: {
    duration?: number;
    easing?: EasingFunction;
    interpolate?: any;
  } = {
    duration: 25,
  };
  const tween = tweened(value, tweenProps);
  $: tween.set(value, tweenProps);
</script>

<slot output={$tween} />
