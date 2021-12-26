let queryNumber = 0;

function guidGenerator() {
	var S4 = function () {
		return (((1 + Math.random()) * 0x10000) | 0).toString(16).substring(1);
	};
	return S4() + S4() + '-' + S4() + '-' + S4() + '-' + S4() + '-' + S4() + S4() + S4();
}

interface DataModellerState {
    queries: Query[];
    activeQuery?: string;
}

interface Query {
    query: string;
    name: string;
    id: string;
    cardinality?: number;
    sizeInBytes?: number;
    error?: string;
    profile?: any;
    preview?: any;
    destinationProfile?: any;
}

export function emptyQuery(): Query {
	const id = guidGenerator();
	queryNumber += 1;
	return {
		query: '',
		name: `query_${queryNumber}.sql`,
		id,
        profile: undefined,
        preview: undefined,
        sizeInBytes: undefined
	};
}

export function initialState() : DataModellerState {
    return {
        queries: [emptyQuery()]
    }
}

/**
 * These actions will be threaded into the server storage.
 * Each action function in the object can take one of two forms:
 * 1. a plain function that represents an immer mutation of the state object;
 * 2. an async (function) that has a dispatch and getState function. Think of these
 * as a thunk, like in Redux. The dispatch function takes in the plain immer mutation function.
 * You can access these by marking the action function as async,
 * e.g.
 * {
 *      async updateQuery() {
 *          return (dispatch, getState) => {
 *              dispatch(doSomething());
 *              const state = getState();
                dispatch(draft => { draft.value = 10 });
 *          }
 * }
 * }
 * @returns {object}
 */
export const createServerActions = (api, notifyUser) => {
    return (store, options) => ({
        addQuery() {
            return draft => { 
                    draft.queries.push(emptyQuery()); 
            };
        },
        updateQuery({id, query}) {
            return (draft) => {
                draft.queries.find((q) => q.id === id).query = query;
            };
        },

        setActiveQuery({id}) {
            return (draft:DataModellerState) => {
                draft.activeQuery = id;
            }
        },

        changeQueryName({id, name}) {
            return (draft:DataModellerState) => {
                draft.queries.find((q) => q.id === id).name = name;
            }
        },
        deleteQuery({id}) {
            return (draft:DataModellerState) => {
                draft.queries = draft.queries.filter(q => q.id !== id);
            }
        },

        moveQueryDown({id}) { 
            return (draft:DataModellerState) => {
                const idx = draft.queries.findIndex((q) => q.id === id);
                if (idx < draft.queries.length - 1) {
                    const thisQuery = { ...draft.queries[idx] };
                    const nextQuery = { ...draft.queries[idx + 1] };
                    draft.queries[idx] = nextQuery;
                    draft.queries[idx + 1] = thisQuery;
                }
            };
        },

        moveQueryUp({id}) {
            return (draft:DataModellerState) => {
                const idx = draft.queries.findIndex((q) => q.id === id);
                if (idx > 0) {
                    const thisQuery = { ...draft.queries[idx] };
                    const nextQuery = { ...draft.queries[idx - 1] };
                    draft.queries[idx] = nextQuery;
                    draft.queries[idx - 1] = thisQuery;
                }
            }
        },

        exportToParquet({query, id, path}) {
            return async (dispatch) => {
                await api.exportToParquet(query, path);

                api.getDestinationSize(path).then((size) => {
                    if (size !== undefined) {
                        dispatch((draft:DataModellerState) => {
                            let q = draft.queries.find(query => query.id === id);
                            q.sizeInBytes = size;
                        })
                    }
                })
                notifyUser({ message: `exported ${path}`, type: "info"})

                return true;
            }
        },

        updateQueryInformation({id}) {
            return async (dispatch, getState) => {
                
                const state = getState();
                const queryInfo = state.queries.find(query => query.id === id);
                // check to see if it is valid.
                const checked = await api.checkQuery(queryInfo.query);
                if (checked.status === 'ERROR') {
                    dispatch((draft:DataModellerState) => {
                        let q = draft.queries.find(query => query.id === id);
                        q.error = checked.error;
                    });
                    // return early.
                    return;
                }

                // if valid, wrap query as temp view.
                try {
                    await api.wrapQueryAsView(queryInfo.query);
                } catch (err) {
                    console.log('reached an error', err);
                }
                
                // get the preview dataset.
                api.createPreview(queryInfo.query).then((preview) => {
                    dispatch((draft:DataModellerState) => {
                        let q = draft.queries.find(query => query.id === id);
                        if (preview.error) {
                            //
                        } else {
                            q.preview = preview.results;
                        }
                    });
                })
                /** The source profile */
                // FIXME: work with parquet files only?
                // api.createSourceProfile(queryInfo.query).then((profile) => {
                //     console.log('created source profile', profile);
                //     dispatch((draft) => {
                //         let q = draft.queries.find(query => query.id === id);
                //         q.profile = profile;
                //     });
                // })
                api.createSourceProfileFromParquet(queryInfo.query).then((profile) => {
                    dispatch((draft:DataModellerState) => {
                        let q = draft.queries.find(query => query.id === id);
                        q.profile = profile;
                    });
                })

                api.calculateDestinationCardinality(queryInfo.query).then((cardinality) => {
                    dispatch((draft:DataModellerState) => {
                        let q = draft.queries.find(query => query.id === id);
                        q.cardinality = cardinality;
                    });
                })

                api.getDestinationSize(`./export/${queryInfo.name.replace('.sql', '.parquet')}`).then((size) => {
                    if (size !== undefined) {
                        dispatch((draft:DataModellerState)=> {
                            let q = draft.queries.find(query => query.id === id);
                            q.sizeInBytes = size;
                        })
                    }
                })

                api.createDestinationProfile(queryInfo.query).then((tableInfo) => {
                    dispatch((draft:DataModellerState) => {
                        let q = draft.queries.find(query => query.id === id);
                        q.destinationProfile = tableInfo;
                    })
                })

            }
        }
    })
}