<!--
@component
A convenience function that takes a value, and through slot props exposes a tweened version
of the value.
-->
<script lang="ts">
  import { tweened } from "svelte/motion";
  export let value: any;
  export let tweenProps = { duration: 25 };
  const tween = tweened(value, tweenProps);
  $: tween.set(value);
</script>

<slot output={$tween} />
