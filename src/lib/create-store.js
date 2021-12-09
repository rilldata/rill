import { writable, get } from 'svelte/store';
//import { produce } from "immer/dist/immer.cjs.production.min.js";
import { produce } from 'immer';

/*
My state approach.

an initializer itself is just a function that returns the initial state.
a plugin must return an object w/ the key nextStore
The main premise
*/

export function initializeFromLocalStorage(key) {
	return (initialState) => {
		const value = localStorage.getItem(key);
		if (value !== null) return JSON.parse(value);
		return initialState;
	};
}

export function saveToLocalStorage(key) {
	return (store) => {
		setInterval(() => {
			localStorage.setItem(key, JSON.stringify(get(store)));
		}, 500);
		return { nextStore: store };
	};
}

export function addProduce() {
	return (store) => ({
		nextStore: store,
		produce(fcn) {
			store.update(
				produce((draft) => {
					fcn(draft);
				})
			);
		},
		setField(key, value) {
			this.produce((draft) => {
				draft[key] = value;
			});
		}
	});
}

export function timeTravel(length = 100) {
	return (store) => {
		let stack = [];
		let index = 0;
		let wasHistory = true;
		store.subscribe(($store) => {
			if (wasHistory) {
				// reset if index was rewound, cut future tape.
				if (index < stack.length - 1) {
					stack = stack.slice(0, index + 1);
				}
				// push to stack.
				stack.push($store);
				// remove old tape at end of reel.
				if (stack.length > length) {
					stack = stack.slice(stack.length - length);
				}
				index = stack.length - 1;
			} else {
				wasHistory = true;
			}
		});
		function undo() {
			if (index > 0) {
				index -= 1;
				wasHistory = false;
				store.set(stack[index]);
			}
		}
		function redo() {
			if (index < length - 1 && index < stack.length - 1) {
				index += 1;
				wasHistory = false;
				store.set(stack[index]);
			}
		}

		return {
			nextStore: store,
			undo,
			redo
		};
	};
}

export function loggable(store) {
	store.subscribe(console.log);
	return { nextStore: store };
}

export function resettable(initialState) {
	return (nextStore) => {
		return {
			nextStore,
			reset() {
				nextStore.set(initialState);
			}
		};
	};
}

export function withPlugins(...pluginSet) {
	return (store, initialState) => {
		return pluginSet.reduce(
			([nextStore, options], plugin) => {
				let nextOptions = options;
				if (plugin) {
					nextOptions = plugin(nextStore, initialState);
					// continuously appends new key value pairs to the store.
					options = { ...options, ...nextOptions };
				}
				return [nextOptions.nextStore, options];
			},
			[store, {}]
		);
	};
}

export function createStore(initialState, ...plugins) {
	const initialStore = writable(initialState);
	const [store, etc] = withPlugins(...plugins)(initialStore, initialState);
	return {
		subscribe: store.subscribe,
		...etc
	};
}

// EXAMPLE:
// export function everything(initialState, localStorageKey) {
//   return createStore(
//     initializeFromLocalStorage(localStorageKey)(initialState),
//     addProduce(),
//     timeTravel(),
//     resettable(initialState),
//     saveToLocalStorage(localStorageKey),
//     loggable
//   );
// }
