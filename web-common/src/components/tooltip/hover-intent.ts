import type { Action } from "svelte/action";

interface HoverIntentParams {
  threshold?: number;
  timeout?: number;
  activeDelay?: number;
  hideDelay?: number;
  onActiveChange?: (active: boolean) => void;
}

export const hoverIntent: Action<HTMLElement, HoverIntentParams> = (
  node,
  params = {},
) => {
  const {
    threshold = 5,
    timeout = 100,
    activeDelay = 200,
    hideDelay = 0,
    onActiveChange = () => {},
  } = params;

  let isHovering = false;
  let isMouseMoved = false;
  let isPointerMoving = false;
  let lastMouseX = 0;
  let lastMouseY = 0;

  let waitUntilTimer: ReturnType<typeof setTimeout> | undefined;
  let hoverIntentTimer: ReturnType<typeof setTimeout> | undefined;
  let resetMoveTimer: ReturnType<typeof setTimeout> | undefined;
  let moveThrottleTimer: ReturnType<typeof setTimeout> | undefined;

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
    if (moveThrottleTimer) {
      clearTimeout(moveThrottleTimer);
      moveThrottleTimer = undefined;
    }
  }

  function waitUntil(callback: () => void, time = activeDelay) {
    clearAllTimers();
    waitUntilTimer = setTimeout(callback, time);
  }

  function resetMoveState() {
    isMouseMoved = false;
    isPointerMoving = false;

    if (isHovering) {
      hoverIntentTimer = setTimeout(() => {
        if (!isMouseMoved && isHovering) {
          waitUntil(() => {
            onActiveChange(true);
          });
        }
      }, timeout);
    }
  }

  function handlePointerEnter(event: PointerEvent) {
    isHovering = true;
    lastMouseX = event.clientX;
    lastMouseY = event.clientY;
    isMouseMoved = false;
    isPointerMoving = false;

    clearAllTimers();

    hoverIntentTimer = setTimeout(() => {
      if (!isMouseMoved && isHovering) {
        waitUntil(() => {
          onActiveChange(true);
        });
      }
    }, timeout);
  }

  function handlePointerMove(event: PointerEvent) {
    if (!isHovering || isPointerMoving) return;

    isPointerMoving = true;
    moveThrottleTimer = setTimeout(() => {
      isPointerMoving = false;
    }, 16); // ~60fps

    const deltaX = Math.abs(event.clientX - lastMouseX);
    const deltaY = Math.abs(event.clientY - lastMouseY);

    if (deltaX > threshold || deltaY > threshold) {
      isMouseMoved = true;
      clearAllTimers();
      onActiveChange(false);

      resetMoveTimer = setTimeout(resetMoveState, timeout);
    }

    lastMouseX = event.clientX;
    lastMouseY = event.clientY;
  }

  function handlePointerLeave() {
    isHovering = false;
    isMouseMoved = false;
    isPointerMoving = false;
    clearAllTimers();

    waitUntil(() => {
      onActiveChange(false);
    }, hideDelay);
  }

  node.addEventListener("pointerenter", handlePointerEnter);
  node.addEventListener("pointermove", handlePointerMove);
  node.addEventListener("pointerleave", handlePointerLeave);

  return {
    update(newParams: HoverIntentParams) {
      Object.assign(params, newParams);
    },
    destroy() {
      clearAllTimers();
      node.removeEventListener("pointerenter", handlePointerEnter);
      node.removeEventListener("pointermove", handlePointerMove);
      node.removeEventListener("pointerleave", handlePointerLeave);
    },
  };
};
