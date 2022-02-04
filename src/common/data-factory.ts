import type {DataModelerState, Dataset, Model} from "$lib/types";
import {guidGenerator} from "$lib/util/guid";
import {sanitizeQuery as _sanitizeQuery} from "$lib/util/sanitize-query";

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

export function newQuery(params: NewQueryArguments = {}): Model {
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