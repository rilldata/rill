/**
 * NOTE: only pure JS functions are allowed here. Anything that requires implementations
 * should be assumed to exist in the api object passed to createServerActions.
 * This enables us to swap out different APIs & backends as needed.
 */
import { createDatasetActions } from "./dataset/index.js";
import { createModelActions } from "./model/index.js";
import { createMetricsModelActions } from "./metrics-model/index.js";
import { createExploreConfigurationActions } from "./explore/index.js";
import type { Item, Model, Dataset, MetricsModel, DataModelerState } from "../lib/types"
import { guidGenerator } from "../lib/util/guid.js";
import { sanitizeQuery as _sanitizeQuery } from "../lib/util/sanitize-query.js";

let queryNumber = 0;

interface NewQueryArguments { 
    query?: string;
    name?: string;
    at?: number;
}

export function newSource(): Dataset {
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

export function newQuery(params:NewQueryArguments = {}): Model {
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

export function emptyQuery(): Model {
	return newQuery({});
}

export function initialState() : DataModelerState {
    return {
        queries: [emptyQuery()],
        sources: [],
        metricsModels: [],
        exploreConfigurations: [],
        status: 'disconnected'
    }
}

function getByID(items:(Item[]), id:string) : Item| null {
    return items.find(q => q.id === id);
}

export function addError(dispatch:Function, id:string, message:string) : void {
    dispatch((draft:DataModelerState) => {
        let q = getByID(draft.queries, id) as Model;
        q.error = message;
    });
}

function clearQuery(dispatch:Function, id:string) : void {
    dispatch((draft:DataModelerState) => {
        let q = getByID(draft.queries, id) as Model;
        q.sizeInBytes = undefined;
        q.destinationProfile = undefined;
        q.preview = undefined;
        q.profile = undefined;
    });
}

function clearError(dispatch:Function, id:string) {
    dispatch((draft:DataModelerState) => {
        let q =  getByID(draft.queries, id) as Model;
        q.error = undefined;
    });
}

function sanitizeQuery(dispatch:Function, id:string) {
    dispatch((draft:DataModelerState) => {
        let q =  getByID(draft.queries, id) as Model;
        q.sanitizedQuery = _sanitizeQuery(q.query);
    });
}

function updateQueryField(dispatch:Function, id:string, field:string, value:any) {
    dispatch((draft:DataModelerState) => {
        let q = getByID(draft.queries, id);
        q[field] = undefined;
        q[field] = value;
    });
}

const debounceTimers = {};
function debounce(timerKey, fcn, timeout = 500) {
    if (debounceTimers[timerKey]) clearTimeout(debounceTimers[timerKey]);
    debounceTimers[timerKey] = setTimeout(() => {
      fcn();
    }, timeout)
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
export const createDataModelerActions = (api, notifyUser) => {
    return (store, options) => ({
        // sources
        setDBStatus(state:string) {
            return (draft:DataModelerState) => {
                draft.status = state;
            }
        },

        ...createDatasetActions(api),
        ...createModelActions(api),
        // we won't develop these two action sets too much going forward.
        ...createMetricsModelActions(api),
        ...createExploreConfigurationActions(api),

        // FIXME: should this move to src/server/dataset/index.ts?
        // FIXME: rename source => dataset

        computeModelProfile({ id }) {
            return async (dispatch:Function, getState:() => DataModelerState) => {
                //get query
                const state = getState();
                const query = getByID(state.queries, id) as Model;
                const tableName = query.name.split(".sql")[0];
                // update destinationPreview
                // the destinationPreview shoudl be of type Source
                // get this table's fields.
                const path = `./export/${tableName}.parquet`;
                
                // drop the 
                dispatch((draft:DataModelerState) => {
                    const profile = (getByID(draft.queries, id) as Model).profile;
                    if (profile) {
                        profile.forEach(field => {
                            field.summary = undefined;
                            field.nullCount = undefined;
                        })
                    }
                })
                // let's do it. the hard thing. let's materialize the query.
                //notifyUser({ message: `materializing ${tableName}`, type: "info"})
                try {
                    console.time(`materialize: ${tableName}`)
                    //await api.materializeTable(tableName, query.query);
                    await api.createViewOfQuery(tableName, query.query)
                    console.timeEnd(`materialize: ${tableName}`)
                } catch (err) {
                    console.log('we are hitting this error state')
                    console.log(err);
                }
                
                const profile = await api.createDestinationProfile(tableName);
                dispatch((draft:DataModelerState) => {
                    (getByID(draft.queries, id) as Model).profile = profile;
                })
                dispatch((draft:DataModelerState) => {
                    const sourceToUpdate = getByID(draft.queries, id) as Dataset;
                    sourceToUpdate.profile.map((t) => {
                        sourceToUpdate.profile.find(p => p.name === t.name).conceptualType = t.type;
                    });
                })
                profile.forEach(field => {
                    if (field.type === 'VARCHAR') {
                        dispatch(this.summarizeCategoricalField(query.id, tableName, field.name, 'queries'));
                        
                    } else {
                        dispatch(this.summarizeNumericField(query.id, tableName, field.name, field.type, 'queries'));
                        if (field.type === 'TIMESTAMP') {
                            dispatch(this.summarizeTimestampRange(query.id, tableName, field.name, 'queries'));
                        }
                    }
                    dispatch(this.summarizeNullCount(query.id, tableName, field.name, 'queries'));
                })
            };
        },

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
                        const sourceToUpdate = getByID(draft.sources, source.id) as Dataset;
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
                            dispatch(this.summarizeTimestampRange(source.id, parquetPath, field.name));
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

        exportToParquet({query, id, path}) {
            return async (dispatch:Function) => {
                await api.exportToParquet(query, path);

                api.getDestinationSize(path).then((size) => {
                    if (size !== undefined) {
                        dispatch((draft:DataModelerState) => {
                            let q = draft.queries.find(query => query.id === id);
                            q.sizeInBytes = size;
                        })
                    }
                });
                notifyUser({ message: `exported ${path}`, type: "info"});
                return true;
            }
        },

        updateQueryInformation({id}) {
            return async (dispatch:Function, getState:Function) => {
                const state = getState();
                const queryInfo = state.queries.find(query => query.id === id);

                // STEP ONE
                // check to see if it is valid.
                try {
                    await api.validateQuery(queryInfo.query);
                } catch (error) {
                    if (error.message !== 'No statement to prepare!') {
                        addError(dispatch, id, error.message);
                    }  else {
                        clearQuery(dispatch, id);
                    } 
                    return;
                }
                // reset 
                clearError(dispatch, id);

                // let's check if the query differs from the last sanitized queyr.
                const sanitized = queryInfo.sanitizedQuery;
                const thisQuery = queryInfo.query;
                const nextSanitizedQuery = _sanitizeQuery(thisQuery);

                // if they are not the same, let's debounce a re-materialization of the fields.
                if (sanitized !== nextSanitizedQuery) {
                    debounce('destination-profile', () => {
                        const state = getState();
                        const queryInfo = state.queries.find(query => query.id === id);
                        if (queryInfo) {
                            dispatch(this.computeModelProfile({ id: queryInfo.id }))
                        } else {
                            console.info('model removed before we could debounce');
                        }
                    }, 1000);
                } else {
                    return;
                }
                


                sanitizeQuery(dispatch, id);
                
                const tableName = queryInfo.name.split('.sql')[0];

                // if valid, wrap query as temp view.
                try {
                    await api.wrapQueryAsView(queryInfo.query, 'tmp');
                } catch (err) {
                    console.error('reached an error', err);
                }
                
                let anyRemainingErrors = false;
                // get the preview dataset.

                // Check for groupBy in this query.
                // if group by exists, debounce.

                const hasGroupBy = nextSanitizedQuery.includes('group by');

                // only generate preview every 500ms;

                function preview() {
                    api.getPreviewDataset(queryInfo.query, 'tmp').then((preview) => {
                        updateQueryField(dispatch, id, 'preview', preview);
                    }).catch(error => {
                        console.error('createPreview', error);
                    });
                }
                if (hasGroupBy) {
                    debounce('has-group-by', preview, 200)
                } else {
                    preview();
                }
                

                // FIXME: we need to generalize this source table crawl.
                const sources = api.extractParquetFilesFromQuery(queryInfo.query);
                updateQueryField(dispatch, id, 'sources', sources);

                api.getTransformRowCardinality(queryInfo.query, 'tmp').then((cardinality) => {
                    updateQueryField(dispatch, id, 'cardinality', cardinality);
                }).catch(error => {
                    console.error('calculateDestinationCardinality', error);
                });

                api.getDestinationSize(`./export/${queryInfo.name.replace('.sql', '.parquet')}`)
                    .then((size:number) => {
                        if (size !== undefined) {
                            updateQueryField(dispatch, id, 'sizeInBytes', size);
                        }
                    }).catch(error => {
                        console.error('getDestinationSize', error);
                    });

                api.createDestinationProfile('tmp').then((destinationProfile) => {
                    updateQueryField(dispatch, id, 'destinationProfile', destinationProfile);
                }).catch(error => {
                    console.error('createDestinationProfile', error);
                });



            }
        }
    })
}