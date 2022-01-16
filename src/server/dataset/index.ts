/**
 * dataset.ts
 * contains the actions that can be taken to construct a dataset.
 */

import type { DataModelerState, Source, Item } from "src/types"

export function getByID(items:(Item[]), id:string) : Item| null {
    return items.find(q => q.id === id);
}

export async function calculateFieldProperty(
    howToUpdate,
    apiFcn,
    dispatch,
    getState,
    datasetID,
    tableOrPath,
    field,
    ...args:any[]
) {
    // check to see if this field 
    const state = getState();
    // check to see if the function has been called before.
    const targetSource = getByID(state.sources, datasetID) as Source;
    const profileField = targetSource.profile.find(({ name }) => name === field);
    if (!('summary' in profileField)) {
        apiFcn(tableOrPath, field, ...args).then((summary) => {
            dispatch((draft:DataModelerState) => {
                const sourceToUpdate = getByID(draft.sources, datasetID) as Source;
                const profile = sourceToUpdate.profile.find(p => p.name === field);
                howToUpdate(profile, summary);
            })
        })
    }
}

export async function summarizeCategoricalField(
        /** the ID of the datase as expressed in DataModelerState.sources. */
        datasetID:string, 
        /** the table or path to source. THis often appears in the FROM statement. */
        tableOrPath:string, 
        /** the field name, aka the column to analyze. */
        field:string, 
        /** the API object that contains the getTopKAndCardinality implementation.
         * We separate the API call itself from this thunk function so we can swap out implementations
         * (e.g. different sql engines + test mocks)
         */
        api, 
        /** this function needs to check on the initial state first before making choices */
        getState: () => DataModelerState,
        /** the dispatch function that gets fed into immer's produce. */ 
        dispatch:Function
    ) {
    return calculateFieldProperty(
        (profile, summary) => {
            profile.summary = summary;
        },
        api.getTopKAndCardinality,
        dispatch,
        getState,
        datasetID,
        tableOrPath,
        field
    );
}

export async function summarizeNumericField(
    sourceID:string,
    tableOrPath:string,
    field:string,
    fieldType:string,
    api,
    getState: () => DataModelerState,
    dispatch:Function
) {
    return calculateFieldProperty(
        (profile, histogram) => {
            if (!('summary'in profile)) {
                    profile.summary = {};
                }
            profile.summary.histogram = histogram;
        },
        api.numericHistogram,
        dispatch,
        getState,
        sourceID,
        tableOrPath,
        field,
        fieldType
    )
}

export async function summarizeNullCounts(
    datasetID:string,
    tableOrPath:string,
    field:string,
    api,
    getState: () => DataModelerState,
    dispatch:Function
) {
    return calculateFieldProperty(
        (profile, nullCount) => {
            profile.nullCount = nullCount;
        },
        api.getNullCount,
        dispatch,
        getState,
        datasetID,
        tableOrPath,
        field
    )
}