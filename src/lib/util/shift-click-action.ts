import { setContext } from "svelte";
import transientBooleanStore from "$lib/util/transient-boolean-store"

interface CreateShiftClick {
    stopImmediatePropagation: boolean
}

export function createShiftClickAction(params : CreateShiftClick = { stopImmediatePropagation: true}) {
    let _stopImmediatePropagation = params?.stopImmediatePropagation || false;
	const clickSwitch = transientBooleanStore();
	// set a context for children to consume transient state.
	setContext("rill:app:ui:shift-click", clickSwitch);
	
	return {
		// export the click switch store if needed in the attached component
		clickSwitch,
		// put this in a use:shiftClickAction
		shiftClickAction(node:Element) {
			function shiftClick(event:MouseEvent) {
				if (event.shiftKey) {
						// dispatch our custom event. accessible via on:shift-click
						node.dispatchEvent(new CustomEvent("shift-click"));
						// update our shared store.
						clickSwitch.flip();
						// prevent the regular on:click event here.
						if (_stopImmediatePropagation) {
							event.stopImmediatePropagation();
						}
				}
			}
			node.addEventListener('click', shiftClick);
			return {
				destroy() {
					// remove the shiftClick listener if the DOM element
					node.addEventListener('click', shiftClick);
				}
			}
		}
	}
}