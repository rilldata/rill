/**
 * @module mousePositionToDomainActionFactory
 * This action factory creates
 * 1. a readable store that contains the domain coordinates
 * 2. an action that updates the readable store when the mouse moves over the attached DOM element
 */
import { getContext } from "svelte";
import { get, Readable, writable } from "svelte/store";
import { DEFAULT_COORDINATES } from "../constants";

import type { Action, ActionReturn } from "svelte/action";
import { contexts } from "../constants";
import type { DomainCoordinates } from "../constants/types";
import type { ScaleStore } from "../state/types";

export interface MousePositionToDomainActionSet {
  coordinates: Readable<DomainCoordinates>;
  mousePositionToDomain: Action<HTMLElement | SVGElement>;
}

export function mousePositionToDomainActionFactory(): MousePositionToDomainActionSet {
  const coordinateStore = writable<DomainCoordinates>({
    ...DEFAULT_COORDINATES,
  });
  const xScale = getContext(contexts.scale("x")) as ScaleStore;
  const yScale = getContext(contexts.scale("y")) as ScaleStore;

  let offsetX: number;
  let offsetY: number;
  let mouseover = false;

  const unsubscribeFromXScale = xScale.subscribe((xs) => {
    if (mouseover) {
      coordinateStore.update((coords) => {
        return { ...coords, x: xs(offsetX) };
      });
    }
  });
  const unsubscribeFromYScale = yScale.subscribe((ys) => {
    if (mouseover) {
      coordinateStore.update((coords) => {
        return { ...coords, y: ys(offsetY) };
      });
    }
  });

  function onMouseMove(event) {
    /** fixme: use  */
    offsetX = event.offsetX;
    offsetY = event.offsetY;

    coordinateStore.set({
      x: get(xScale).invert(offsetX),
      y: get(yScale).invert(offsetY),
    });
    mouseover = true;
  }

  function onMouseLeave() {
    coordinateStore.set({ ...DEFAULT_COORDINATES });
    mouseover = false;
  }
  const coordinates = {
    subscribe: coordinateStore.subscribe,
  } as Readable<DomainCoordinates>;
  return {
    coordinates,
    mousePositionToDomain(node: HTMLElement | SVGElement): ActionReturn<void> {
      node.addEventListener("mousemove", onMouseMove);
      node.addEventListener("mouseleave", onMouseLeave);
      return {
        destroy(): void {
          unsubscribeFromXScale();
          unsubscribeFromYScale();
          node.removeEventListener("mousemove", onMouseMove);
          node.removeEventListener("mouseleave", onMouseLeave);
        },
      };
    },
  };
}
