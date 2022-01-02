/**
 * NOTE: only pure JS functions are allowed here. Anything that requires implementations
 * should be assumed to exist in the api object passed to createServerActions.
 * This enables us to swap out different APIs & backends as needed.
 */
import { sanitizeQuery as _sanitizeQuery } from "../util/sanitize-query.js";

let queryNumber = 0;

function guidGenerator() {
	var S4 = function () {
		return (((1 + Math.random()) * 0x10000) | 0).toString(16).substring(1);
	};
	return S4() + S4() + '-' + S4() + '-' + S4() + '-' + S4() + '-' + S4() + S4() + S4();
}

interface DataModellerState {
    activeQuery?: string;
    queries: Query[];
    sources: Source[];
}

interface NewQueryArguments {
    query?: string;
    name?: string;
}

interface Query {
    query: string;
    sanitizedQuery: string;
    name: string;
    id: string;
    cardinality?: number;
    sizeInBytes?: number; // TODO: make sure this is just size
    error?: string;
    profile?: ProfileColumn[]; // TODO: create Profile interface
    preview?: any;
    destinationProfile?: any;
}

interface Source {
    id: string;
    path: string;
    name: string;
    profile: ProfileColumn[]; // TODO: create Profile interface
    head: any[];
    cardinality?: number;
    sizeInBytes?: number;
}

/**
 * The type definition for a "profile column"
 */
interface ProfileColumn {
    name: string;
    type: string;
}

export function newSource(): Source {
    return {
        id: guidGenerator(),
        path: '',
        name: '',
        profile: [],
        cardinality: undefined,
        sizeInBytes: undefined,
        head: []
    }
}

export function newQuery(params:NewQueryArguments = {}): Query {
    const query = params.query || '';
    const sanitizedQuery = _sanitizeQuery(query);
    const name = `${params.name || `query_${queryNumber}`}.sql`;
    queryNumber += 1;
    return {
		query,
        sanitizedQuery,
		name,
		id: guidGenerator(),
        profile: undefined,
        preview: undefined,
        sizeInBytes: undefined
	};
}

export function emptyQuery(): Query {
	return newQuery({});
}

export function initialState() : DataModellerState {
    return {
        queries: [emptyQuery()],
        sources: []
    }
}

function getQuery(queries, id) {
    return queries.find(q=>q.id === id);
}

function addError(dispatch:Function, id:string, message:string) {
    dispatch((draft:DataModellerState) => {
        let q = getQuery(draft.queries, id);
        q.error = message;
    });
}

function clearQuery(dispatch:Function, id:string) {
    dispatch((draft:DataModellerState) => {
        let q = getQuery(draft.queries, id);
        q.sizeInBytes = undefined;
        q.destinationProfile = undefined;
        q.preview = undefined;
        q.profile = undefined;
    });
}

function clearError(dispatch:Function, id:string) {
    dispatch((draft:DataModellerState) => {
        let q =  getQuery(draft.queries, id);
        q.error = undefined;
    });
}

function sanitizeQuery(dispatch:Function, id:string) {
    dispatch((draft:DataModellerState) => {
        let q =  getQuery(draft.queries, id);
        q.sanitizedQuery = _sanitizeQuery(q.query);
    });
}

function updateQueryField(dispatch:Function, id:string, field:string, value:any) {
    dispatch((draft:DataModellerState) => {
        let q = getQuery(draft.queries, id);
        q[field] = value;
    });
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
        // sources
        clearSources() {
            return (draft:DataModellerState) => {
                draft.sources = [];
            }
        },
        addOrUpdateSource(path) {
            return async (dispatch:Function, getState:Function) => {
                const sources = getState().sources;
                const sourceExists = sources.find(s => s.path === path);
                const source = {...(sourceExists || newSource())};
                source.path = path;
                source.name = path.split('/').slice(-1)[0];
                try {
                    source.profile = await api.createSourceProfile(source.path);
                    source.profile = source.profile.filter(row => row.name !== 'duckdb_schema');
                    source.sizeInBytes = await api.getDestinationSize(source.path);
                    source.cardinality = await api.getCardinality(source.path);
                    source.head = await api.getFirstN(`'${source.path}'`);
                    dispatch((draft:DataModellerState) => {
                        if (!!sourceExists) {
                            const sourceToUpdate = getQuery(draft.sources, source.id);
                            Object.keys(source).forEach((k) => {
                                sourceToUpdate[k] = source[k];
                            })
                        } else {
                            draft.sources.push(source);
                        }
                    })
                } catch (err) {
                    console.log("addSource", err, path);
                    //throw Error(err);
                }
            }
        },

        scanRootForSources() {
            return async (dispatch, getState) => {
                const files = await api.getParquetFilesInRoot();
                files.sort();
                const filePaths = new Set(files);
                // prune
                dispatch((draft:DataModellerState) => {
                    draft.sources = draft.sources.filter(s => filePaths.has(s.path));
                })
                files.forEach(path => {
                    try {
                        dispatch(this.addOrUpdateSource(path))
                    } catch (err) {
                        console.log(err, path);
                    }
                });
            }
        },

        // queries
        addQuery(params) {
            const query = params.query || undefined;
            const name = params.name || undefined;
            const at = params.at;
            return (draft:DataModellerState) => {
                if (at !== undefined) {
                    draft.queries = [...draft.queries.slice(0, at), newQuery({ query, name }), ...draft.queries.slice(at)];
                } else {
                    draft.queries.push(newQuery({ query, name })); 
                }
            };
        },
        updateQuery({id, query}) {
            return (draft:DataModellerState) => {
                getQuery(draft.queries, id).query = query;
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
                });
                notifyUser({ message: `exported ${path}`, type: "info"});
                //dispatch(this.scanRootForSources());
                return true;
            }
        },

        updateQueryInformation({id}) {
            return async (dispatch, getState) => {
                const state = getState();
                const queryInfo = state.queries.find(query => query.id === id);
                // check to see if it is valid.
                try {
                    await api.checkQuery(queryInfo.query);
                } catch (error) {
                    if (error.message !== 'No statement to prepare!') {
                        console.log(id);
                        addError(dispatch, id, error.message);
                    }  else {
                        clearQuery(dispatch, id);
                    } 
                    return;
                }
                // reset 
                clearError(dispatch, id);
                sanitizeQuery(dispatch, id);

                // if valid, wrap query as temp view.
                try {
                    await api.wrapQueryAsView(queryInfo.query);
                } catch (err) {
                    console.error('reached an error', err);
                }
                
                let anyRemainingErrors = false;
                // get the preview dataset.
                api.createPreview(queryInfo.query).then((preview) => {
                    updateQueryField(dispatch, id, 'preview', preview);
                }).catch(error => {
                    console.error('createPreview', error);
                });

                api.createSourceProfileFromQuery(queryInfo.query).then((profile) => {
                    updateQueryField(dispatch, id, 'profile', profile);
                }).catch(error => {
                    console.error('createSourceProfile', error);
                    addError(dispatch, id, error.message);
                });

                api.calculateDestinationCardinality(queryInfo.query).then((cardinality) => {
                    updateQueryField(dispatch, id, 'cardinality', cardinality);
                }).catch(error => {
                    console.error('calculateDestinationCardinality', error);
                });

                api.getDestinationSize(`./export/${queryInfo.name.replace('.sql', '.parquet')}`)
                    .then((size) => {
                        if (size !== undefined) {
                            updateQueryField(dispatch, id, 'sizeInBytes', size);
                        }
                    }).catch(error => {
                        console.error('getDestinationSize', error);
                    });

                api.createDestinationProfile(queryInfo.query).then((destinationProfile) => {
                    updateQueryField(dispatch, id, 'destinationProfile', destinationProfile);
                }).catch(error => {
                    console.error('createDestinationProfile', error);
                });

            }
        }
    })
}