<!-- @component 
measures the bounding rect of this element's child
-->
<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  export let tag = "div";

  let element: HTMLElement | SVGElement;
  let observer;
  let rect;
  let styles;

  function toNumber(px) {
    if (!px) return 0;
    return px?.includes("px") ? +px.split("px")[0] : 0;
  }
  onMount(() => {
    observer = new ResizeObserver(() => {
      rect = element.parentElement.getBoundingClientRect();
      styles = window.getComputedStyle(element.parentElement);
    });
    observer.observe(element.parentElement);
    rect = element.parentElement.getBoundingClientRect();
  });

  onDestroy(() => {
    observer.unobserve(element.parentElement);
  });
</script>

<svelte:element this={tag} bind:this={element} class="contents">
  <slot {rect} {styles} {toNumber} />
</svelte:element>
