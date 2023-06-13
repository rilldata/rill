<script lang="ts">
  import { onMount } from "svelte";

  export let text: string = undefined;
  let element;

  onMount(() => {
    const line = element?.parentNode?.parentNode;
    // move this DOM element to the right side of the line?
  });
</script>

<div class="float-right text-red-300">
  <span bind:this={element} class="inline-block outline-black">
    {text}
  </span>
</div>
