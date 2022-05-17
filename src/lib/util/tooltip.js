// eslint-disable-next-line import/no-extraneous-dependencies
import { listen } from "svelte/internal";
import { placeElement } from "../utils/float-placement";

const defaults = {
  duration: 50,
  location: "bottom",
  alignment: "center",
  distance: 4,
  visible: true,
};

export function tooltip(node, args) {
  const options = { ...defaults, ...args };
  const { duration, location, alignment, distance } = options;
  const el = document.createElement("div");
  el.className = "tooltip";
  el.textContent = options.text;
  el.style.position = "absolute";
  el.style.transition = `opacity ${duration}ms, transform ${duration}ms`;

  let { visible } = options;
  let entered = false;

  function setLocation() {
    const [left, top] = placeElement({
      location,
      alignment,
      distance,
      parentPosition: node.getBoundingClientRect(),
      elementPosition: el.getBoundingClientRect(),
      y: window.scrollY,
    });

    el.style.top = `${top}px`;
    el.style.left = `${left}px`;
  }

  function updateElement(newText, isVisible) {
    el.textContent = newText;
    document.body.appendChild(el);
    el.style.opacity = "0";
    if (newText && isVisible && entered) {
      setTimeout(() => {
        el.style.visibility = "visible";
        el.style.opacity = "1";
        visible = true;
      });
    } else {
      setTimeout(() => {
        el.style.opacity = "0";
        el.style.visibility = "hidden";
        // visible = false;
      });
    }
    setLocation();
  }

  function append() {
    entered = true;
    if (el.textContent.length && options.text) {
      updateElement(options.text, visible);
    }
  }

  function remove() {
    entered = false;
    el.remove();
  }

  const removeEnter = listen(node, "mouseenter", append);
  const removeLeave = listen(node, "mouseleave", remove);

  return {
    destroy() {
      remove();
      removeEnter();
      removeLeave();
    },
    update(newArgs) {
      visible = newArgs.visible === undefined || newArgs.visible;
      updateElement(newArgs.text, visible);
    },
  };
}
