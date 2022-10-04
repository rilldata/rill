<script lang="ts">
  import { createEventDispatcher, setContext } from "svelte";
  import { cubicOut } from "svelte/easing";
  import { tweened } from "svelte/motion";
  import { derived, writable } from "svelte/store";

  const selectedValue = writable(undefined);
  const dispatch = createEventDispatcher();
  function callback(element, value) {
    dispatch("select", value);
    selectedValue.set(element);
  }

  setContext("rill:app:tabgroup-callback", callback);
  setContext("rill:app:tabgroup-selected", selectedValue);

  let movingElementRect = derived(selectedValue, ($element) => {
    const r = $element?.getBoundingClientRect();
    if (r) return { width: r.width, left: r.left, top: r.bottom };
  });

  let tweenedMovingElement = tweened($movingElementRect, {
    duration: 200,
    easing: cubicOut,
  });
  $: tweenedMovingElement.set($movingElementRect);
</script>

<div class="flex flex-row gap-x-4">
  <slot />
  {#if $selectedValue !== undefined && $tweenedMovingElement?.left}
    <div
      class="absolute rounded bg-gray-600"
      style:left="{$tweenedMovingElement.left}px"
      style:top="calc({$tweenedMovingElement.top}px - .25rem)"
      style:width="{$tweenedMovingElement.width}px"
      style:height=".25rem"
    />
  {/if}
</div>
