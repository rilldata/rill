<!-- @component 
measures the bounding rect of this element's child
-->
<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  export let tag = "div";

  let element: HTMLElement | SVGElement;
  let observer;
  let rect;
  onMount(() => {
    console.log(element);
    observer = new ResizeObserver(() => {
      rect = element.parentElement.getBoundingClientRect();
    });
    observer.observe(element.parentElement);
    rect = element.parentElement.getBoundingClientRect();
  });

  onDestroy(() => {
    observer.unobserve(element.parentElement);
  });
</script>

<svelte:element this={tag} bind:this={element} class="contents">
  <slot {rect} />
</svelte:element>
