<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  let container;
  let pxWidth: number;

  let observer: ResizeObserver;

  const measureSpan = () => {
    let rect = container.getBoundingClientRect();
    // pxWidth = Math.ceil(rect.right - rect.left);
    pxWidth = rect.right - rect.left;
    console.log({ pxWidth });
  };

  onMount(() => {
    const observer = new ResizeObserver(() => {
      measureSpan();
    });
    // measureSpan();
    observer.observe(container);

    document.fonts.onloadingdone = () => {
      measureSpan();
    };
  });

  onDestroy(() => {
    observer?.disconnect();
  });
</script>

<div>
  <span bind:this={container} style="outline: solid 1px #ccc;"><slot /></span>
  <span> -- {pxWidth}px</span>
</div>
