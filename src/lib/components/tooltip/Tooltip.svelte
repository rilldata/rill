<script lang="ts">
import { onMount } from "svelte";
import { fade } from "svelte/transition"
import { placeElement } from '$lib/util/place-element';

export let location = 'bottom';
export let alignment = 'middle';
export let distance = 0;
/** the delay in miliseconds before rendering the tooltip once mouse has entered */
export let activeDelay = 150;
/** the delay in miliseconds before unrendering the tooltip once mouse has left */
export let nonActiveDelay = 0;

export let active = false;

let top = 0;
let left = 0;
let innerHeight;
let innerWidth;
let scrollY;
let scrollX;

function setLocation(parentBoundingClientRect, elementBoundingClientRect, scrollXValue, scrollYvalue, windowWidth, windowHeight) {
    if (!(parentBoundingClientRect && elementBoundingClientRect)) return;
    const [leftPos, topPos] = placeElement({
      location,
      alignment,
      distance,
      parentPosition: parentBoundingClientRect,
      elementPosition: elementBoundingClientRect,
      y: scrollYvalue,
      x: scrollXValue,
      windowWidth, windowHeight
    });
    top = topPos;
    left = leftPos;
  }
let parent;
let child;

let waitUntilTimer;
function waitUntil(callback, time = 150) {
    if (waitUntilTimer) clearTimeout(waitUntilTimer);
    waitUntilTimer = setTimeout(callback, time);
}

// $: firstParentElement = parent?.children[0].getBoundingClientRect();
let firstParentElement;
$: if (firstParentElement) setLocation(firstParentElement.getBoundingClientRect(), child?.getBoundingClientRect(), scrollX, scrollY, innerWidth, innerHeight);

onMount(() => {
    // we listen to the parent.
    // actually, we listen to the first chidl element! 
    firstParentElement = parent?.children[0];
    const config = { attributes: true  };

    const observer = new MutationObserver(() => {
        setLocation(firstParentElement.getBoundingClientRect(), child?.getBoundingClientRect(), scrollX, scrollY, innerWidth, innerHeight);
    })
    observer.observe(firstParentElement, config);


})

</script>

<svelte:window bind:scrollX bind:scrollY bind:innerHeight bind:innerWidth />

<div class='fixed bg-black text-white p-3' style:left=0px style:top=0px>
    {scrollX} <b>innerWidth</b> {innerWidth} <b>innerHeight</b> {innerHeight}
</div>

<div class='contents' bind:this={parent}  
    on:mouseenter={() => { 
        waitUntil(() => { 
            active = true;
        }, activeDelay); 
    }} 
    on:mouseleave={() => { 
        waitUntil(() => {
            active = false;
        }, nonActiveDelay);
    }}>
    <slot />
    {#if active}
        <div transition:fade|local={{duration: 50 }} bind:this={child} class="absolute" style:left="{left}px" style:top="{top}px">
            <slot name="tooltip-content" />
        </div>
    {/if}
</div>
