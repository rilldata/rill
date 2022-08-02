<script lang="ts">
  import {
    mouseLocationToBoundingRect,
    placeElement,
  } from "$lib/util/place-element";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import Portal from "../Portal.svelte";

  export let target;
  export let relationship = "parent"; // parent, mouse {x, y}
  export let location = "bottom";
  export let alignment = "middle";
  export let distance = 0;
  // edge padding
  export let pad = 8;
  /** the delay in miliseconds before rendering the tooltip once mouse has entered */
  /** the delay in miliseconds before unrendering the tooltip once mouse has left */

  let top = 0;
  let left = 0;
  let innerHeight;
  let innerWidth;
  let scrollY;
  let scrollX;

  function setLocation(
    parentBoundingClientRect,
    elementBoundingClientRect,
    scrollXValue,
    scrollYvalue,
    windowWidth,
    windowHeight
  ) {
    if (!(parentBoundingClientRect && elementBoundingClientRect)) return;
    const [leftPos, topPos] = placeElement({
      location,
      alignment,
      distance,
      pad,
      parentPosition: parentBoundingClientRect,
      elementPosition: elementBoundingClientRect,
      y: scrollYvalue,
      x: scrollXValue,
      windowWidth,
      windowHeight,
    });
    top = topPos;
    left = leftPos;
  }
  let child;

  let firstParentElement;

  function getFirstValidChildElement(element) {
    // get this child.
    let possibleChild = element?.children[0];
    // check for display: contents, which may indicate
    // another wrapped object.
    console.log(possibleChild, getComputedStyle(possibleChild).display);
    if (getComputedStyle(possibleChild).display === "contents") {
      return getFirstValidChildElement(possibleChild);
    } else {
      return possibleChild;
    }
  }

  $: if (relationship === "parent") {
    if (firstParentElement)
      setLocation(
        firstParentElement.getBoundingClientRect(),
        child?.getBoundingClientRect(),
        scrollX,
        scrollY,
        innerWidth,
        innerHeight
      );
  } else {
    setLocation(
      relationship === "direct"
        ? target.getBoundingClientRect()
        : mouseLocationToBoundingRect(target),
      child?.getBoundingClientRect(),
      scrollX,
      scrollY,
      innerWidth,
      innerHeight
    );
  }
  $: getFirstValidChildElement(target);

  onMount(() => {
    // we listen to the parent.
    // actually, we listen to the first chidl element!
    if (relationship === "parent") {
      firstParentElement = getFirstValidChildElement(target); // target?.children[0];
      const config = { attributes: true };
      const observer = new MutationObserver(() => {
        setLocation(
          firstParentElement.getBoundingClientRect(),
          child?.getBoundingClientRect(),
          scrollX,
          scrollY,
          innerWidth,
          innerHeight
        );
      });
      if (firstParentElement) {
        observer.observe(firstParentElement, config);
      }
    }
  });
</script>

<svelte:window bind:scrollX bind:scrollY bind:innerHeight bind:innerWidth />

<Portal>
  <div
    transition:fade|local={{ duration: 25 }}
    bind:this={child}
    class="absolute"
    style:z-index="200"
    style:left="{left}px"
    style:top="{top}px"
  >
    <slot />
  </div>
</Portal>
