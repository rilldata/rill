import { setContext } from "svelte";
import { get, writable } from "svelte/store";

interface CreateCommandClick {
  stopImmediatePropagation: boolean;
}

export function createCommandClickAction(
  params: CreateCommandClick = { stopImmediatePropagation: true },
) {
  const _stopImmediatePropagation = params?.stopImmediatePropagation || false;
  // set a context for children to consume transient state.
  const { subscribe, update } = writable([]);

  const commandHeld = writable(false);
  // create a callback store that can be added to by children components.
  // see StackingWord.svelte for an example of usage.
  const callbacks = {
    subscribe,
    addCallback(callback) {
      update((cbs) => {
        return [...cbs, callback];
      });
    },
  };

  setContext("rill:app:ui:command-click-action-callbacks", callbacks);

  return {
    // export the click switch callbacks added by children, in case that's needed.
    clickCommandCallbacks: callbacks,
    commandHeld,
    // put this in a use:commandClickAction
    commandClickAction(node: Element) {
      function mouseDown(event) {
        if (event.ctrlKey || event.metaKey) {
          commandHeld.set(true);
          node.dispatchEvent(new CustomEvent("command-mousedown"));
        }
      }

      function mouseUp(event) {
        commandHeld.set(false);
        if (event.ctrlKey || event.metaKey)
          node.dispatchEvent(new CustomEvent("command-mouseup"));
      }
      function commandClick(event: MouseEvent) {
        if (event.ctrlKey || event.metaKey) {
          // dispatch our custom event. accessible via on:command-click
          node.dispatchEvent(new CustomEvent("command-click"));
          // fire all callbacks.
          const cbs = get(callbacks);
          cbs.forEach((cb: () => void) => cb());

          // prevent the regular on:click event here.
          if (_stopImmediatePropagation) {
            event.stopImmediatePropagation();
          }
        }
        node.addEventListener("mousedown", mouseDown);
      }
      node.addEventListener("mouseup", mouseUp);
      node.addEventListener("click", commandClick);
      window.addEventListener("mouseup", mouseUp);
      return {
        destroy() {
          node.removeEventListener("mousedown", mouseDown);
          node.removeEventListener("mouseup", mouseUp);
          window.removeEventListener("mouseup", mouseUp);
          node.removeEventListener("click", commandClick);
        },
      };
    },
  };
}
