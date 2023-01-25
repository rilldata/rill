<script lang="ts">
  import { onMount } from "svelte";

  export let string: string;
  let stringContainer;
  let pxWidth: number;

  onMount(() => {
    let rect = stringContainer.getBoundingClientRect();
    pxWidth = Math.ceil(rect.right - rect.left);
  });
</script>

<div>
  <span bind:this={stringContainer} style="outline: solid 1px #ccc;"
    >{string}</span
  >
  <span> -- {pxWidth}px</span>
</div>
