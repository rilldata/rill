import {
	createStore,
	initializeFromLocalStorage,
	saveToLocalStorage,
	loggable,
	resettable,
	addProduce
} from './create-store';

function guidGenerator() {
	var S4 = function () {
		return (((1 + Math.random()) * 0x10000) | 0).toString(16).substring(1);
	};
	return S4() + S4() + '-' + S4() + '-' + S4() + '-' + S4() + '-' + S4() + S4() + S4();
}

let queryNumber = 0;
function newQuery() {
	const id = guidGenerator();
	queryNumber += 1;
	return {
		query: '',
		name: `query_${queryNumber}.sql`,
		id
	};
}

const generateInitialState = () => ({
	queries: [newQuery()]
});

export function initialize() {
	const initialState = generateInitialState();
	const store = createStore(
		//initialState,
		initializeFromLocalStorage('app')(initialState),
		saveToLocalStorage('app'),
		addProduce(),
		//loggable,
		resettable(initialState)
	);

	store.createQuery = () =>
		store.produce((draft) => {
			draft.queries.push(newQuery());
		});

	store.changeQueryName = (qid, name) => {
		store.produce((draft) => {
			draft.queries.find((q) => q.id === qid).name = name;
		});
	};

	store.editQuery = (qid, newQuery) =>
		store.produce((draft) => {
			draft.queries.find((q) => q.id === qid).query = newQuery;
		});

	store.deleteQuery = (qid) =>
		store.produce((draft) => {
			draft.queries = draft.queries.filter((q) => q.id !== qid);
		});

	store.moveQueryDown = (qid) => {
		store.produce((draft) => {
			const idx = draft.queries.findIndex((q) => q.id === qid);
			if (idx < draft.queries.length - 1) {
				const thisQuery = { ...draft.queries[idx] };
				const nextQuery = { ...draft.queries[idx + 1] };
				draft.queries[idx] = nextQuery;
				draft.queries[idx + 1] = thisQuery;
			}
		});
	};

	store.moveQueryUp = (qid) => {
		store.produce((draft) => {
			const idx = draft.queries.findIndex((q) => q.id === qid);
			if (idx > 0) {
				const thisQuery = { ...draft.queries[idx] };
				const nextQuery = { ...draft.queries[idx - 1] };
				draft.queries[idx] = nextQuery;
				draft.queries[idx - 1] = thisQuery;
			}
		});
	};

	return store;
}
