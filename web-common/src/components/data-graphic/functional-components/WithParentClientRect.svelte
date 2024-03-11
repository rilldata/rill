<!-- @component 
measures the bounding rect of this element's child
-->
<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  export let tag = "div";

  let element: HTMLElement | SVGElement;
  let observer: ResizeObserver;
  let rect: DOMRect;
  let styles: CSSStyleDeclaration | null;
  let parentElement: HTMLElement | null;

  function toNumber(px: string | undefined) {
    if (!px) return 0;
    return px?.includes("px") ? +px.split("px")[0] : 0;
  }
  onMount(() => {
    parentElement = element.parentElement;

    if (!parentElement) return;

    observer = new ResizeObserver((entries) => {
      const entry = entries[0];

      rect = entry.contentRect;
      styles =
        element?.parentElement &&
        window.getComputedStyle(element?.parentElement);
    });

    observer.observe(parentElement);
  });

  onDestroy(() => {
    if (!parentElement) return;
    observer.unobserve(parentElement);
  });
</script>

<svelte:element this={tag} bind:this={element} class="contents">
  <slot {rect} {styles} {toNumber} />
</svelte:element>
