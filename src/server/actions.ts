/**
 * NOTE: only pure JS functions are allowed here. Anything that requires implementations
 * should be assumed to exist in the api object passed to createServerActions.
 * This enables us to swap out different APIs & backends as needed.
 */
import { createDatasetActions } from "./dataset/index.js";
import { createTransformActions } from "./transform/index.js";
import type { Item, Query, Source, DataModelerState } from "../types"
import { guidGenerator } from "../util/guid.js";
import { sanitizeQuery as _sanitizeQuery } from "../util/sanitize-query.js";

let queryNumber = 0;

interface NewQueryArguments { 
    query?: string;
    name?: string;
    at?: number;
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
        preview: undefined,
        sizeInBytes: undefined
	};
}

export function emptyQuery(): Query {
	return newQuery({});
}

export function initialState() : DataModelerState {
    return {
        queries: [emptyQuery()],
        sources: [],
        status: 'disconnected'
    }
}

function getByID(items:(Item[]), id:string) : Item| null {
    return items.find(q => q.id === id);
}

function addError(dispatch:Function, id:string, message:string) : void {
    dispatch((draft:DataModelerState) => {
        let q = getByID(draft.queries, id) as Query;
        q.error = message;
    });
}

function clearQuery(dispatch:Function, id:string) : void {
    dispatch((draft:DataModelerState) => {
        let q = getByID(draft.queries, id) as Query;
        q.sizeInBytes = undefined;
        q.destinationProfile = undefined;
        q.preview = undefined;
        q.profile = undefined;
    });
}

function clearError(dispatch:Function, id:string) {
    dispatch((draft:DataModelerState) => {
        let q =  getByID(draft.queries, id) as Query;
        q.error = undefined;
    });
}

function sanitizeQuery(dispatch:Function, id:string) {
    dispatch((draft:DataModelerState) => {
        let q =  getByID(draft.queries, id) as Query;
        q.sanitizedQuery = _sanitizeQuery(q.query);
    });
}

function updateQueryField(dispatch:Function, id:string, field:string, value:any) {
    dispatch((draft:DataModelerState) => {
        let q = getByID(draft.queries, id);
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
        setDBStatus(state:string) {
            return (draft:DataModelerState) => {
                draft.status = state;
            }
        },

        ...createDatasetActions(api),
        ...createTransformActions(api),

        // FIXME: should this move to src/server/dataset/index.ts?
        // FIXME: rename source => dataset
        addOrUpdateSource(path) {
            return async (dispatch:Function, getState:()=>DataModelerState) => {
                const sources = getState().sources;
                const sourceExists = sources.find(s => s.path === path);
                const source = {...(sourceExists || newSource())};
                source.path = path;
                source.name = path.split('/').slice(-1)[0];
                try {
                    if (!('profile' in source && source.profile.length)) {
                        source.profile = await api.createSourceProfile(source.path);
                        source.profile = source.profile.filter(row => row.name !== 'duckdb_schema' && row.name !== 'schema');
                    }
                    source.sizeInBytes = await api.getDestinationSize(source.path);
                    source.cardinality = await api.getCardinality(source.path);
                    source.head = await api.getFirstN(`'${source.path}'`);
                    dispatch((draft:DataModelerState) => {
                        if (!!sourceExists) {
                            const sourceToUpdate = getByID(draft.sources, source.id);
                            // replace 
                            Object.keys(source).forEach((k) => {
                                sourceToUpdate[k] = source[k];
                            })
                        } else {
                            draft.sources.push(source);
                        }
                    });

                    const duckdbTypes = await api.parquetToDBTypes(source.path);
                    dispatch((draft:DataModelerState) => {
                        const sourceToUpdate = getByID(draft.sources, source.id) as Source;
                        duckdbTypes.map((t) => {
                            sourceToUpdate.profile.find(p => p.name === t.name).conceptualType = t.type;
                        });
                    })
                    // run expensive stuff but update it async.
                    const numerics = duckdbTypes.filter(c => {
                        return c.type.includes("INTEGER") || c.type.includes("DOUBLE") || c.type.includes("BIGINT");
                    });
                    const strings = duckdbTypes.filter(c => {
                        return c.type.includes("VARCHAR");
                    });

                    const timestamps = duckdbTypes.filter(c => {
                        return c.type.includes('TIMESTAMP');
                    });

                    const parquetPath = `'${source.path}'`;

                    if (strings.length) {
                        strings.forEach(field => {
                            dispatch(this.summarizeCategoricalField(source.id, parquetPath, field.name));
                        });
                    }

                    if (numerics.length) {
                        numerics.forEach((field) => {
                            dispatch(this.summarizeNumericField(source.id, parquetPath, field.name, field.type));
                        })
                    }

                    if (timestamps.length) {
                        timestamps.forEach(field => {
                            dispatch(this.summarizeNumericField(source.id, parquetPath, field.name, field.type));
                        })
                    }
                    duckdbTypes.forEach(field => {
                        dispatch(this.summarizeNullCount(source.id, parquetPath, field.name));
                    })
                    
                } catch (err) {
                    console.log("addSource", err, path);
                }
            }
        },

        scanRootForSources() {
            return async (dispatch:Function) => {
                const files = await api.getParquetFilesInRoot();
                files.sort();
                const filePaths = new Set(files);
                // prune & dedup
                dispatch((draft:DataModelerState) => {
                    draft.sources = draft.sources.filter(s => filePaths.has(s.path));
                    draft.sources = draft.sources.filter((value, index, self) =>
                        index === self.findIndex((t) => (t.path === value.path))
                    );
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
        // addQuery(params:NewQueryArguments) {
        //     const query = params.query || undefined;
        //     const name = params.name || undefined;
        //     const at = params.at;
        //     return (draft:DataModelerState) => {
        //         if (at !== undefined) {
        //             draft.queries = [...draft.queries.slice(0, at), newQuery({ query, name }), ...draft.queries.slice(at)];
        //         } else {
        //             draft.queries.push(newQuery({ query, name })); 
        //         }
        //     };
        // },
        // updateQuery({id, query}) {
        //     return (draft:DataModelerState) => {
        //         const queryItem = getByID(draft.queries, id) as Query;
        //         queryItem.query = query;
        //     };
        // },

        // setActiveQuery({id}) {
        //     return (draft:DataModelerState) => {
        //         draft.activeQuery = id;
        //     }
        // },

        // changeQueryName({id, name}) {
        //     return (draft:DataModelerState) => {
        //         draft.queries.find((q) => q.id === id).name = name;
        //     }
        // },
        // deleteQuery({id}) {
        //     return (draft:DataModelerState) => {
        //         draft.queries = draft.queries.filter(q => q.id !== id);
        //     }
        // },

        // moveQueryDown({id}) { 
        //     return (draft:DataModelerState) => {
        //         const idx = draft.queries.findIndex((q) => q.id === id);
        //         if (idx < draft.queries.length - 1) {
        //             const thisQuery = { ...draft.queries[idx] };
        //             const nextQuery = { ...draft.queries[idx + 1] };
        //             draft.queries[idx] = nextQuery;
        //             draft.queries[idx + 1] = thisQuery;
        //         }
        //     };
        // },

        // moveQueryUp({id}) {
        //     return (draft:DataModelerState) => {
        //         const idx = draft.queries.findIndex((q) => q.id === id);
        //         if (idx > 0) {
        //             const thisQuery = { ...draft.queries[idx] };
        //             const nextQuery = { ...draft.queries[idx - 1] };
        //             draft.queries[idx] = nextQuery;
        //             draft.queries[idx - 1] = thisQuery;
        //         }
        //     }
        // },

        // exportToParquet({query, id, path}) {
        //     return async (dispatch:Function) => {
        //         await api.exportToParquet(query, path);

        //         api.getDestinationSize(path).then((size) => {
        //             if (size !== undefined) {
        //                 dispatch((draft:DataModelerState) => {
        //                     let q = draft.queries.find(query => query.id === id);
        //                     q.sizeInBytes = size;
        //                 })
        //             }
        //         });
        //         notifyUser({ message: `exported ${path}`, type: "info"});
        //         return true;
        //     }
        // },

        // updateQueryInformation({id}) {
        //     return async (dispatch:Function, getState:Function) => {
        //         const state = getState();
        //         const queryInfo = state.queries.find(query => query.id === id);
        //         // check to see if it is valid.
        //         try {
        //             await api.checkQuery(queryInfo.query);
        //         } catch (error) {
        //             if (error.message !== 'No statement to prepare!') {
        //                 console.log(id);
        //                 addError(dispatch, id, error.message);
        //             }  else {
        //                 clearQuery(dispatch, id);
        //             } 
        //             return;
        //         }
        //         // reset 
        //         clearError(dispatch, id);
        //         sanitizeQuery(dispatch, id);

        //         // if valid, wrap query as temp view.
        //         try {
        //             await api.wrapQueryAsView(queryInfo.query);
        //         } catch (err) {
        //             console.error('reached an error', err);
        //         }
                
        //         let anyRemainingErrors = false;
        //         // get the preview dataset.
        //         api.createPreview(queryInfo.query).then((preview) => {
        //             updateQueryField(dispatch, id, 'preview', preview);
        //         }).catch(error => {
        //             console.error('createPreview', error);
        //         });

        //         // attach the source name here.
        //         const sources = api.extractParquetFilesFromQuery(queryInfo.query);
        //         updateQueryField(dispatch, id, 'sources', sources);



        //         api.calculateDestinationCardinality(queryInfo.query).then((cardinality) => {
        //             updateQueryField(dispatch, id, 'cardinality', cardinality);
        //         }).catch(error => {
        //             console.error('calculateDestinationCardinality', error);
        //         });

        //         api.getDestinationSize(`./export/${queryInfo.name.replace('.sql', '.parquet')}`)
        //             .then((size:number) => {
        //                 if (size !== undefined) {
        //                     updateQueryField(dispatch, id, 'sizeInBytes', size);
        //                 }
        //             }).catch(error => {
        //                 console.error('getDestinationSize', error);
        //             });

        //         api.createDestinationProfile(queryInfo.query).then((destinationProfile) => {
        //             updateQueryField(dispatch, id, 'destinationProfile', destinationProfile);
        //         }).catch(error => {
        //             console.error('createDestinationProfile', error);
        //         });

        //     }
        // }
    })
}