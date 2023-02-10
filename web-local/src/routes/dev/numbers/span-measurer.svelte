<script lang="ts">
  import { onMount } from "svelte";

  let container;
  let pxWidth: number;

  onMount(() => {
    let rect = container.getBoundingClientRect();
    // pxWidth = Math.ceil(rect.right - rect.left);
    pxWidth = rect.right - rect.left;
  });
</script>

<div>
  <span bind:this={container} style="outline: solid 1px #ccc;"><slot /></span>
  <span> -- {pxWidth}px</span>
</div>
