<script lang="ts">
  import { tweened } from "svelte/motion";

  export let isCandidate = false;

  let heightTween = tweened(0, { duration: 0 });
  $: heightTween.set(isCandidate ? 1 : 0, {
    duration: 100,
  });
</script>

<div
  on:mouseover
  on:focus
  on:mouseleave
  style:height={"38px"}
  style:left="0px"
  class="z-40 w-full -translate-y-[19px]"
  class:absolute={!isCandidate}
  class:static={isCandidate}
  style:outline="4px solid red"
/>
