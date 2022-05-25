import { writable, get } from "svelte/store";
import { DEFAULT_COORDINATES } from "./constants";

/** converts an event to a simplified object
 * with only the needed properties
 */
function mouseEvents(event: MouseEvent) {
  return {
    movementX: event.movementX,
    movementY: event.movementY,
    clientX: event.clientX,
    clientY: event.clientY,
  };
}

interface ScrubActionFactoryArguments {
  /** the bounds where the scrub is active. */
  plotLeft: number;
  plotRight: number;
  plotTop: number;
  plotBottom: number;
  /** the name of the events we declare for start, move, end.
   * Typically mousedown, mousemove, and mouseup.
   */
  startEvent?: string;
  endEvent?: string;
  moveEvent?: string;
  /** the dispatched move event name for the scrub move effect, to be
   * passed up to the parent element when the scrub move has happened.
   * e.g.
   */
  moveEventName?: string;
  /** the dispatched move event name for the scrub completion effect, to be
   * passed up to the parent element when the scrub is completed.
   * e.g.
   */
  completedEventName?: string;
  /** These predicates will gate whether we continue with
   * the startEvent, moveEvent, and endEvents.
   * If they're not passed in as arguments, the action
   * will always assume they're true.
   * This is used e.g. when a user wants to hold the shift or alt key, or 
   * check for some other condition to to be true.
   */
  startPredicate?: (event: Event) => boolean;
  movePredicate?: (event: Event) => boolean;
  endPredicate?: (event: Event) => boolean;
}

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
}: ScrubActionFactoryArguments) {
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
