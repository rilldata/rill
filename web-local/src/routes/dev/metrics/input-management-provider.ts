import { setContext } from "svelte";
import { get, Writable, writable } from "svelte/store";

/** a provider function that enables downstream components to add their corresponding element to the list of DOMElements
 * in order
 */
export function createInputManagementProvider() {
  const domElements: Writable<HTMLElement[]> = writable([]);

  function blurInputs() {
    get(domElements).forEach((element) => {
      element.blur();
    });
  }

  function elementInFocus() {
    return get(domElements).some(
      (element) => document.activeElement === element
    );
  }

  setContext("rill:app:blurInputs", blurInputs);
  setContext("rill:app:inputDOMElements", domElements);
  return { blurInputs, elementInFocus };
}
