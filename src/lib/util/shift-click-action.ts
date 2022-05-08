import { setContext } from "svelte";
import { writable, get } from 'svelte/store';

interface CreateShiftClick {
    stopImmediatePropagation: boolean
}

export function createShiftClickAction(params : CreateShiftClick = { stopImmediatePropagation: true}) {
    let _stopImmediatePropagation = params?.stopImmediatePropagation || false;
	// set a context for children to consume transient state.
	let { subscribe, update } = writable([]);
	
	let shiftHeld = writable(false);
	// create a callback store that can be added to by children components.
	// see StackingWord.svelte for an example of usage.
	let callbacks = {
			subscribe,
			addCallback(callback) {
				update(cbs => {
					return [...cbs, callback]
			})
		}
	}

	setContext("rill:app:ui:shift-click-action-callbacks", callbacks);
	
	return {
		// export the click switch callbacks added by children, in case that's needed.
		clickShiftCallbacks: callbacks,
		shiftHeld,
		// put this in a use:shiftClickAction
		shiftClickAction(node:Element) {
			function mouseDown(event) {
				if (event.shiftKey) {
					shiftHeld.set(true);
					node.dispatchEvent(new CustomEvent("shift-mousedown"));
				}
			}

			function mouseUp(event) {
				shiftHeld.set(false);
				if (event.shiftKey) node.dispatchEvent(new CustomEvent("shift-mouseup"));
			}
			function shiftClick(event:MouseEvent) {
				if (event.shiftKey) {
						// dispatch our custom event. accessible via on:shift-click
						node.dispatchEvent(new CustomEvent("shift-click"));
						// fire all callbacks.
						const cbs = get(callbacks);
						cbs.forEach((cb) => cb());

						// prevent the regular on:click event here.
						if (_stopImmediatePropagation) {
							event.stopImmediatePropagation();
						}
				}
			node.addEventListener('mousedown', mouseDown);}
			node.addEventListener('mouseup', mouseUp);
			node.addEventListener('click', shiftClick);
			window.addEventListener('mouseup', mouseUp);
			return {
				destroy() {
					node.removeEventListener('mousedown', mouseDown);
					node.removeEventListener('mouseup', mouseUp);
					window.removeEventListener('mouseup', mouseUp);
					node.removeEventListener('click', shiftClick);
				}
			}
		}
	}
}