import type { SvelteComponent } from "svelte";
import Tooltip from "./Tip.svelte";
import type { Shortcut } from "./Tip.svelte";

type Options = {
  title?: string;
  text?: string;
  position?: "top" | "bottom" | "left" | "right";
  alignment?: "start" | "center" | "end";
  shortcuts: Shortcut[];
};

export function tooltip(
  element: HTMLElement,
  options: Options = {
    text: "Hellotherherehr",
    position: "bottom",
    alignment: "end",
    shortcuts: [],
  },
) {
  let tooltipComponent: Tooltip;

  function mouseOver() {
    const { width, height, top, left } = element.getBoundingClientRect();

    const midX = left + width / 2;
    const midY = top + height / 2;

    let x = left;
    let y = top;

    switch (options.position) {
      case "top":
        y = top;
        switch (options.alignment) {
          case "start":
            x = left;
            break;
          case "center":
            x = midX;
            break;
          case "end":
            x = left + width;
            break;
        }
        break;
      case "bottom":
        y = top + height;
        switch (options.alignment) {
          case "start":
            x = left;
            break;
          case "center":
            x = midX;
            break;
          case "end":
            x = left + width;
            break;
        }
        break;
      case "left":
        x = left;
        switch (options.alignment) {
          case "start":
            y = top;
            break;
          case "center":
            y = midY;
            break;
          case "end":
            y = top + height;
            break;
        }
        break;
      case "right":
        x = left + width;
        switch (options.alignment) {
          case "start":
            y = top;
            break;
          case "center":
            y = midY;
            break;
          case "end":
            y = top + height;
            break;
        }
        break;
    }

    tooltipComponent = new Tooltip({
      props: {
        title: options.title ?? "",
        text: options.text ?? "",
        x,
        y,
        alignment: options.alignment,
        position: options.position,
        shortcuts: options.shortcuts,
      },
      target: document.body,
    });
  }

  function mouseLeave() {
    console.log("mouseLeave");
    tooltipComponent.$destroy();
  }

  function updateProps(newOptions: Options) {
    if (!tooltipComponent) return;
    tooltipComponent.$set(newOptions);
  }

  element.addEventListener("mouseenter", mouseOver);
  element.addEventListener("mouseleave", mouseLeave);

  return {
    update(newOptions: Options) {
      updateProps(newOptions);
    },
    destroy() {
      element.removeEventListener("mouseover", mouseOver);
      element.removeEventListener("mouseleave", mouseLeave);
    },
  };
}
