<script lang="ts">
import { fly } from "svelte/transition"
import { placeElement } from '$lib/util/place-element';

export let location = 'bottom';
export let alignment = 'right';
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

function setLocation(parentBoundingClientRect, elementBoundingClientRect, scrollYValue, windowWidth, windowHeight) {
    if (!(parentBoundingClientRect && elementBoundingClientRect)) return;
    const [leftPos, topPos] = placeElement({
      location,
      alignment,
      distance,
      parentPosition: parentBoundingClientRect,
      elementPosition: elementBoundingClientRect,
      y: scrollYValue,
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


$: firstElement = parent?.children[0].getBoundingClientRect();

$: setLocation(firstElement, child?.getBoundingClientRect(), scrollY, innerWidth, innerHeight);

</script>

<svelte:window bind:scrollY bind:innerHeight bind:innerWidth />

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
    <div transition:fly|local={{duration: 50  }} bind:this={child} class="absolute" style:left="{left}px" style:top="{top}px">
        <slot name="tooltip-content" />
    </div>
    {/if}
</div>
