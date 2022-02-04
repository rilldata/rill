/**
 * dataset.ts
 * contains the actions that can be taken to construct a dataset.
 */

import type { DataModelerState, Dataset, Item } from "../../lib/types"

// TODO: we use this in other modules. Probably should have single source
export function getByID(items:(Item[]), id:string) : Item| null {
    return items.find(q => q.id === id);
}

/**
 * NOTE: there's some amount of duplication within many of the summarizing functions.
 */
export function createDatasetActions(api) {

    return {

        clearSources() {
            return (draft:DataModelerState) => {
                draft.sources = [];
            }
        },

        /**
         * @summary runs categorical calculations on the field for a given dataset (top-k, cardinality)
         * and updates the store
         * @param datasetID the ID of the dataset
         * @param tableOrPath a table-like used in our FROM statements in API calls
         * @param field the column to summarize
         * @returns 
         */
        summarizeCategoricalField(datasetID:string, tableOrPath:string, field:string, key='sources'){
            return async (dispatch:Function, getState:()=>DataModelerState) => {
                const state = getState();

                const targetSource = getByID(state[key], datasetID) as Dataset;
                const profileField = targetSource.profile.find(({ name }) => name === field);
                if (!('summary' in profileField)) {

                    api.getTopKAndCardinality(tableOrPath, field).then((summary) => {
                        dispatch((draft:DataModelerState) => {
                            const sourceToUpdate = getByID(draft[key], datasetID) as Dataset;
                            const profile = sourceToUpdate.profile.find(p => p.name === field);
                            profile.summary = summary;
                        })
                    })
                }
            }   
        },

        /**
         * @summary calculates a numeric histogram for field for a given dataset and updates the store
         * @param datasetID the ID of the dataset
         * @param tableOrPath a table-like used in our FROM statements in API calls
         * @param field the column to summarize
         * @param fieldType the type of the column; used to handle TIMESTAMPS differently
         * @returns 
         */
        summarizeNumericField(datasetID:string, tableOrPath:string, field:string, fieldType:string, key='sources'){
            return async (dispatch:Function, getState:()=>DataModelerState) => {
                // check to see if this field 
                const state = getState();
                // check to see if the function has been called before.
                const targetSource = getByID(state[key], datasetID) as Dataset;
                const profileField = targetSource.profile.find(({ name }) => name === field);
                if (!('summary' in profileField)) {
                    // clear!
                    // dispatch((draft) => {
                    //     const sourceToUpdate = getByID(draft[key], datasetID) as Source;
                    //     const profile = sourceToUpdate.profile.find(p => p.name === field);
                    //     profile.summary = {};
                    // })
                //if (true) {
                    api.numericHistogram(tableOrPath, field, fieldType).then((histogram) => {
                        dispatch((draft:DataModelerState) => {
                            const sourceToUpdate = getByID(draft[key], datasetID) as Dataset;
                            const profile = sourceToUpdate.profile.find(p => p.name === field);
                            if (!('summary'in profile)) {
                                profile.summary = {};
                            }
                            profile.summary.histogram = histogram;
                        })
                    })
                    if (fieldType !== 'TIMESTAMP') {
                        api.descriptiveStatistics(tableOrPath, field).then((summaryStatistics) => {
                            dispatch((draft:DataModelerState) => {
                                const sourceToUpdate = getByID(draft[key], datasetID) as Dataset;
                                const profile = sourceToUpdate.profile.find(p => p.name === field);
                                if (!('summary'in profile)) {
                                    profile.summary = {};
                                }
                                profile.summary.statistics = summaryStatistics;
                            })
                            
                        })
                    }
                    
                }
            }   
        },

        /**
         * @summary calculates the null counts for a field of a given dataset and updates the store
         * @param datasetID the ID of the dataset
         * @param tableOrPath a table-like used in our FROM statements in API calls
         * @param field the column to summarize
         * @returns 
         */
        summarizeNullCount(datasetID:string, tableOrPath:string, field:string, key='sources'){
            return async (dispatch:Function, getState:()=>DataModelerState) => {
                // check to see if this field 
                const state = getState();
                // check to see if the function has been called before.
                const targetSource = getByID(state[key], datasetID) as Dataset;
                const profileField = targetSource.profile.find(({ name }) => name === field);
                if (!('nullCount' in profileField)) {
                    api.getNullCount(tableOrPath, field).then((nullCount) => {
                        dispatch((draft:DataModelerState) => {
                            const sourceToUpdate = getByID(draft[key], datasetID) as Dataset;
                            const profile = sourceToUpdate.profile.find(p => p.name === field);
                            profile.nullCount = nullCount;
                        })
                    })
                }
            }   
        },

        summarizeTimestampRange(datasetID:string, tableOrPath:string, field:string, key='sources'){
            return async (dispatch:Function, getState:()=>DataModelerState) => {
                // check to see if this field 
                const state = getState();
                // check to see if the function has been called before.
                const targetSource = getByID(state[key], datasetID) as Dataset;
                const profileField = targetSource.profile.find(({ name }) => name === field);
                if (!('summary' in profileField && 'interval' in profileField.summary)) {
                    api.getTimeRange(tableOrPath, field).then((timeInterval) => {
                        dispatch((draft:DataModelerState) => {
                            const sourceToUpdate = getByID(draft[key], datasetID) as Dataset;
                            const profile = sourceToUpdate.profile.find(p => p.name === field);

                            profile.summary = {...profile.summary || {}, ...timeInterval}
                            //profile.summary.interval = timeInterval;
                        })
                    })
                }
            }   
        },
    }
}