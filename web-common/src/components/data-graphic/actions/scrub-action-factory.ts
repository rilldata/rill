/**
 * scrub-action-factory
 * --------------------
 * This action factory produces an object that contains
 * - a coordinates store, which has the x and y start and stop values
 *   of the in-progress scrub.
 * - an isScrubbing store, which the user can exploit to see if scrubbing is
 *   currently happening
 * - a movement store, which captures the momentum of the scrub.
 * - a customized action
 *
 * Why is this an action factory and not an action? Because we actually want to initialize a bunch
 * of stores that are used throughout the app, which respond to the action's logic automatically,
 * and can thus be consumed within the application without any other explicit call point.
 * This action factory pattern is quite useful in a variety of settings.
 * </script>
 */

import { get, writable } from "svelte/store";
import { DEFAULT_COORDINATES } from "../constants";

/** converts an event to a simplified object
 * with only the needed properties
 */
function mouseEvents(event: MouseEvent) {
  return {
    movementX: event.movementX,
    movementY: event.movementY,
    clientX: event.clientX,
    clientY: event.clientY,
    ctrlKey: event.ctrlKey,
    altKey: event.altKey,
    shiftKey: event.shiftKey,
    metaKey: event.metaKey,
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
  startEventName?: string;
  /** the dispatched move event name for the scrub move effect, to be
   * passed up to the parent element when the scrub move has happened.
   * e.g.
   */
  moveEventName?: string;
  /** the dispatched move event name for the scrub completion effect, to be
   * passed up to the parent element when the scrub is completed.
   * e.g. when moveEventName = "scrubbing", we have <div use:scrubAction on:scrubbing={...} />
   */
  endEventName?: string;
  /** These predicates will gate whether we continue with
   * the startEvent, moveEvent, and endEvents.
   * If they're not passed in as arguments, the action
   * will always assume they're true.
   * This is used e.g. when a user wants to hold the shift or alt key, or
   * check for some other condition to to be true.
   * e.g when completedEventName = "scrub", we have <div use:scrubAction on:scrub={...} />
   */
  startPredicate?: (event: Event) => boolean;
  movePredicate?: (event: Event) => boolean;
  endPredicate?: (event: Event) => boolean;
}

export interface PlotBounds {
  plotLeft?: number;
  plotRight?: number;
  plotTop?: number;
  plotBottom?: number;
}

interface ScrubAction {
  destroy: () => void;
}

function clamp(v: number, min: number, max: number) {
  if (v < min) return min;
  if (v > max) return max;
  return v;
}

export function createScrubAction({
  plotLeft,
  plotRight,
  plotTop,
  plotBottom,
  startEvent = "mousedown",
  startEventName = undefined,
  startPredicate = undefined,
  endEvent = "mouseup",
  endPredicate = undefined,
  moveEvent = "mousemove",
  movePredicate = undefined,
  endEventName = undefined,
  moveEventName = undefined,
}: ScrubActionFactoryArguments) {
  const coordinates = writable({
    start: DEFAULT_COORDINATES,
    stop: DEFAULT_COORDINATES,
  });

  /** local plot bound state */
  let _plotLeft = plotLeft;
  let _plotRight = plotRight;
  let _plotTop = plotTop;
  let _plotBottom = plotBottom;

  const movement = writable({
    xMovement: 0,
    yMovement: 0,
  });

  const isScrubbing = writable(false);

  function setCoordinateBounds(event: MouseEvent) {
    return {
      x: clamp(event.offsetX, _plotLeft, _plotRight),
      y: clamp(event.offsetY, _plotTop, _plotBottom),
    };
  }

  return {
    coordinates,
    isScrubbing,
    movement,
    updatePlotBounds(bounds: PlotBounds) {
      if (bounds.plotLeft) _plotLeft = bounds.plotLeft;
      if (bounds.plotRight) _plotRight = bounds.plotRight;
      if (bounds.plotTop) _plotTop = bounds.plotTop;
      if (bounds.plotBottom) _plotBottom = bounds.plotBottom;
    },
    scrubAction(node: Node): ScrubAction {
      function reset() {
        coordinates.set({
          start: DEFAULT_COORDINATES,
          stop: DEFAULT_COORDINATES,
        });
        isScrubbing.set(false);
      }

      function onScrubStart(event: MouseEvent) {
        event.preventDefault();
        if (!(startPredicate === undefined || startPredicate(event))) {
          return;
        }
        coordinates.set({
          start: setCoordinateBounds(event),
          stop: DEFAULT_COORDINATES,
        });
        isScrubbing.set(true);
        if (startEventName) {
          node.dispatchEvent(
            new CustomEvent(startEventName, {
              detail: {
                ...get(coordinates),
                ...mouseEvents(event),
              },
            })
          );
        }
      }

      function onScrub(event: MouseEvent) {
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

      function onScrubEnd(event: MouseEvent) {
        event.preventDefault();
        if (!(endPredicate === undefined || endPredicate(event))) {
          reset();
          return;
        }
        const coords = get(coordinates);
        if (coords.start.x && coords.stop.x && endEventName) {
          node.dispatchEvent(
            new CustomEvent(endEventName, {
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
      window.addEventListener(endEvent, onScrubEnd);
      window.addEventListener(endEvent, reset);
      return {
        destroy() {
          node.removeEventListener(startEvent, onScrubStart);
          node.removeEventListener(moveEvent, onScrub);
          window.removeEventListener(endEvent, onScrubEnd);
          window.removeEventListener(endEvent, reset);
        },
      };
    },
  };
}
