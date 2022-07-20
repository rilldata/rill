import { writable } from "svelte/store";

/** creates a store, hovered, and an action that updates the hovered store
 * when the user mouses over.
 * This is a fast way to extract the hover state of a DOM element.
 */
export function createHoverStateActionFactory() {
  const hovered = writable(false);
  return {
    hovered,
    captureHoverState(node) {
      const hoverState = (trueOrFalse) => () => hovered.set(trueOrFalse);
      const isHovered = hoverState(true);
      const isNotHovered = hoverState(false);

      node.addEventListener("mouseover", isHovered);
      node.addEventListener("focus", isHovered);
      node.addEventListener("mouseleave", isNotHovered);
      node.addEventListener("blur", isNotHovered);
      return {
        destroy() {
          node.removeEventListener("mouseover", isHovered);
          node.removeEventListener("focus", isHovered);
          node.removeEventListener("mouseleave", isNotHovered);
          node.removeEventListener("blur", isNotHovered);
        },
      };
    },
  };
}
