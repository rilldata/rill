<script lang="ts">
  import {
    createEventDispatcher,
    onDestroy,
    onMount,
    setContext,
  } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { derived, Readable, writable } from "svelte/store";

  export let variant: "panel" | "secondary" = "panel";

  let container;

  const selectedValue = writable(undefined);
  const dispatch = createEventDispatcher();
  function callback(element, value) {
    dispatch("select", value);
    selectedValue.set(element);
  }

  // listen to whatever is the current $selectedValue.

  setContext("rill:app:tabgroup-callback", callback);
  setContext("rill:app:tabgroup-selected", selectedValue);

  let movingElementRect: Readable<{
    width: number;
    left: number;
    top: number;
    height: number;
  }> = derived(selectedValue, ($element, set) => {
    const r = $element?.getBoundingClientRect();
    if (r)
      set({
        width: r.width,
        left: $element?.offsetLeft || 0,
        top: $element?.offsetTop || 0,
        height: r.height,
      });
  });

  let tweenedMovingElement = tweened($movingElementRect, {
    duration: 120,
    easing: cubicOut,
  });
  $: tweenedMovingElement.set($movingElementRect);

  let observer;
  let elemBounds;
  onMount(() => {
    observer = new MutationObserver(() => {
      elemBounds = container.getBoundingClientRect();
    });
    observer.observe(container, { childList: true });
  });

  onDestroy(() => {
    observer.disconnect();
  });
</script>

<div
  class="flex flex-row gap-x-4 relative items-stretch"
  bind:this={container}
  style:height="40px"
>
  <slot />
  {#if $selectedValue !== undefined && $tweenedMovingElement?.left !== undefined}
    <div
      class:opacity-20={variant === "secondary"}
      class="absolute rounded bg-gray-600 z-10 pointer-events-none"
      style:left="{$tweenedMovingElement.left}px"
      style:top={variant === "panel"
        ? `calc(${
            $tweenedMovingElement.top + $tweenedMovingElement.height
          }px - .25rem)`
        : `${$tweenedMovingElement.top}px`}
      style:width="{$tweenedMovingElement.width}px"
      style:height={variant === "panel"
        ? ".25rem"
        : `${$tweenedMovingElement.height}px`}
    />
  {/if}
</div>
