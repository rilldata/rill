import { writable, get } from "svelte/store";
import { DEFAULT_COORDINATES } from "./constants";

function mouseEvents(event) {
  return {
    movementX: event.movementX,
    movementY: event.movementY,
    clientX: event.clientX,
    clientY: event.clientY,
  };
}

const DEFAULTS = {
  startEvent: "mousedown",
  startPredicate: undefined,
  endEvent: "mouseup",
  endPredicate: undefined,
  moveEvent: "mousemove",
  movePredicate: undefined,
  completedEventName: undefined,
  moveEventName: undefined,
};

export function createScrubAction({
  plotLeft,
  plotRight,
  plotTop,
  plotBottom,
  startEvent = "mousedown",
  startPredicate = undefined,
  endEvent = "mouseup",
  endPredicate = undefined,
  moveEvent = "mousemove",
  movePredicate = undefined,
  completedEventName = undefined,
  moveEventName = undefined,
}) {
  const coordinates = writable({
    start: DEFAULT_COORDINATES,
    stop: DEFAULT_COORDINATES,
  });

  const movement = writable({
    xMovement: 0,
    yMovement: 0,
  });

  const isScrubbing = writable(false);

  function clamp(v, min, max) {
    if (v < min) return min;
    if (v > max) return max;
    return v;
  }

  function setCoordinateBounds(event) {
    return {
      x: clamp(event.offsetX, plotLeft, plotRight),
      y: clamp(event.offsetY, plotTop, plotBottom),
    };
  }

  return {
    coordinates,
    isScrubbing,
    movement,
    scrubAction(node) {
      function reset() {
        coordinates.set({
          start: DEFAULT_COORDINATES,
          stop: DEFAULT_COORDINATES,
        });
        isScrubbing.set(false);
      }

      function onScrubStart(event) {
        event.preventDefault();
        if (!(startPredicate === undefined || startPredicate(event))) {
          return;
        }
        coordinates.set({
          start: setCoordinateBounds(event),
          stop: DEFAULT_COORDINATES,
        });
        isScrubbing.set(true);
      }

      function onScrub(event) {
        event.preventDefault();
        const isCurrentlyScrubbing = get(isScrubbing);
        if (!isCurrentlyScrubbing) return;
        if (!(movePredicate === undefined || movePredicate(event))) {
          reset();
          return;
        }
        coordinates.update((coords) => {
          const newCoords = { ...coords };
          newCoords.stop = setCoordinateBounds(event);
          return newCoords;
        });
        const coords = get(coordinates);
        // fire the moveEventName event.
        // e.g. on:scrubbing={(event) => { ... }}
        if (moveEventName) {
          node.dispatchEvent(
            new CustomEvent(moveEventName, {
              detail: {
                ...coords,
                ...mouseEvents(event),
              },
            })
          );
        }
      }

      function onScrubEnd(event) {
        event.preventDefault();
        if (!(endPredicate === undefined || endPredicate(event))) {
          reset();
          return;
        }
        const coords = get(coordinates);
        if (coords.start.x && coords.stop.x && completedEventName) {
          node.dispatchEvent(
            new CustomEvent(completedEventName, {
              detail: {
                ...coords,
                ...mouseEvents(event),
              },
            })
          );
        }
        reset();
      }

      node.addEventListener(startEvent, onScrubStart);
      node.addEventListener(moveEvent, onScrub);
      node.addEventListener(endEvent, onScrubEnd);
      window.addEventListener(endEvent, reset);
      return {
        destroy() {
          node.removeEventListener(startEvent, onScrubStart);
          node.removeEventListener(moveEvent, onScrub);
          node.removeEventListener(endEvent, onScrubEnd);
          window.removeEventListener(endEvent, reset);
        },
      };
    },
  };
}
