import { writable, get } from 'svelte/store';
import { produce } from "immer";
import fs from 'fs';

/*
My state approach.

an initializer itself is just a function that returns the initial state.
a plugin must return an object w/ the key nextStore
The main premise
*/

export function initializeFromSavedState(key) {
	return (initialState) => {
		if (fs.existsSync(`${key}.json`)) {
			try {
				return JSON.parse(fs.readFileSync(`${key}.json`).toString());
			} catch (err) {
				console.log(err);
				console.log("going with clean initial state");
			}	
		}
		return initialState;
	};
}

export function saveToLocalFile(key) {
	return (store) => {
		setInterval(() => {
			fs.writeFileSync(`${key}.json`, JSON.stringify(get(store)));
		}, 500);
		return { nextStore: store };
	};
}

export function addProduce() {
	return (store, _, others) => ({
		nextStore: store,
		produce(fcn) {
			// this works very similar to what you'd expect in a redux setting.
			// eg. dispatch(changeChannel('beta')) should take the changeChannel
			// action, which returns a draft-mutating function to be fed into
			// immer's produce function.
			if (fcn.constructor.name === 'AsyncFunction') {
				// I thought about using func.length (if it has two args, then we are go)
    			// but you may only have one. For now, I think marking a function a async
    			// works.
				fcn(this.produce, () => get(store));
			} else {
				// atomic update (singular state change).
				store.update(draft => produce(draft, fcn));
			}
		},
		setField(key, value) {
			this.produce((draft) => {
				draft[key] = value;
			});
		}
	});
}

export function addActions(actionsObject) {
	return (store, _, storeFunctions) => {
		const actionFunctions = actionsObject();
		const actions = Object.keys(actionFunctions).reduce((obj, actionName) => {
			obj[actionName] = (...args) => {
				storeFunctions.produce(actionFunctions[actionName](...args));
			} 
			return obj;
		}, {});
		return {
			nextStore: store,
			...actions
		}
	}
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
				console.log(initialState)
				nextStore.set(initialState);
			}
		};
	};
}

export function listenForSocketMessages() {
	return (nextStore, _, options) => {
		return {
			nextStore,
			listenForSocketMessages(socket) {
				Object.keys(options).forEach(action => {
					socket.on(action, options[action]);
				})
			}
		}
	}
}

export function connectStateToSocket() {
	return (nextStore, _, options) => {
		return {
			nextStore,
			connectStateToSocket(socket) {
				nextStore.subscribe(state => {
					if (socket) {
						socket.emit('app-state', state);
					} else {
						console.log('socket not initialized yet')
					}
				});
				Object.keys(options).forEach(action => {
					if (action !== 'nextStore' && action !== 'socket') {
						socket.on(action, options[action]);
					}
				})
				options.socket = socket;
			}
		}
	}
}

export function withPlugins(...pluginSet) {
	return (store, initialState) => {
		return pluginSet.reduce(
			([nextStore, options], plugin) => {
				let nextOptions = options;
				if (plugin) {
					nextOptions = plugin(nextStore, initialState, options);
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
		get: () => get(store),
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
