<!-- @component 
The FloatingElement component is the backbone of all of our floating UI element functionality.
It handles the setting of the location of the floating element relative to these possible options, set in the relationship prop:
- a direct DOM element passed in through target through the 'direct' prop
- the first child of target through the "parent" prop
display:contents. This is useful when nesting a floating element within a tooltip.
- a mouse click location through "mouse". This is an {x,y} coordinate that matches where the pointer is.
-->
<script lang="ts">
  import {
    mouseLocationToBoundingRect,
    placeElement,
  } from "../../util/place-element";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import Portal from "../Portal.svelte";
  import type { FloatingElementRelationship } from "./types";

  export let target;
  export let relationship: FloatingElementRelationship = "parent"; // parent, mouse {x, y}
  export let location = "bottom";
  export let alignment = "middle";
  export let distance = 0;
  // edge padding
  export let pad = 8;

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
