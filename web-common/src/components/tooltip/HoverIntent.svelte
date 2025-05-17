<script lang="ts">
  import { onDestroy } from "svelte";

  export let threshold = 5;
  export let timeout = 100;
  export let activeDelay = 200;
  export let nonActiveDelay = 0;
  export let active = false;

  let isHovering = false;
  let mouseMoved = false;
  let lastMouseX = 0;
  let lastMouseY = 0;
  let waitUntilTimer: ReturnType<typeof setTimeout> | undefined;
  let hoverIntentTimer: ReturnType<typeof setTimeout> | undefined;
  let resetMoveTimer: ReturnType<typeof setTimeout> | undefined;

  function clearAllTimers() {
    if (waitUntilTimer) {
      clearTimeout(waitUntilTimer);
      waitUntilTimer = undefined;
    }
    if (hoverIntentTimer) {
      clearTimeout(hoverIntentTimer);
      hoverIntentTimer = undefined;
    }
    if (resetMoveTimer) {
      clearTimeout(resetMoveTimer);
      resetMoveTimer = undefined;
    }
  }

  function waitUntil(callback: () => void, time = activeDelay) {
    clearAllTimers();
    waitUntilTimer = setTimeout(callback, time);
  }

  function resetMoveState() {
    mouseMoved = false;

    if (isHovering) {
      hoverIntentTimer = setTimeout(() => {
        if (!mouseMoved && isHovering) {
          waitUntil(() => {
            active = true;
          });
        }
      }, timeout);
    }
  }

  function handlePointerEnter(event: PointerEvent) {
    isHovering = true;
    lastMouseX = event.clientX;
    lastMouseY = event.clientY;
    mouseMoved = false;

    clearAllTimers();

    hoverIntentTimer = setTimeout(() => {
      if (!mouseMoved && isHovering) {
        waitUntil(() => {
          active = true;
        });
      }
    }, timeout);
  }

  function handlePointerMove(event: PointerEvent) {
    if (!isHovering) return;

    const deltaX = Math.abs(event.clientX - lastMouseX);
    const deltaY = Math.abs(event.clientY - lastMouseY);

    if (deltaX > threshold || deltaY > threshold) {
      mouseMoved = true;
      clearAllTimers();
      active = false;

      resetMoveTimer = setTimeout(resetMoveState, timeout);
    }

    lastMouseX = event.clientX;
    lastMouseY = event.clientY;
  }

  function handlePointerLeave() {
    isHovering = false;
    mouseMoved = false;
    clearAllTimers();
    waitUntil(() => {
      active = false;
    }, nonActiveDelay);
  }

  onDestroy(() => {
    clearAllTimers();
  });
</script>

<div
  class="contents"
  on:pointerenter={handlePointerEnter}
  on:pointermove={handlePointerMove}
  on:pointerleave={handlePointerLeave}
>
  <slot />
</div>
