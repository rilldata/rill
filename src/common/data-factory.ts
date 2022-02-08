import type {DataModelerState, Dataset, Model} from "$lib/types";
import {guidGenerator} from "$lib/util/guid";
import {sanitizeQuery as _sanitizeQuery} from "$lib/util/sanitize-query";
import {IDLE_STATUS} from "$common/constants";
import {sanitizeTableName} from "$lib/util/sanitize-table-name";

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
        tableName: '',
        profile: [],
        cardinality: undefined,
        sizeInBytes: undefined,
        head: [],
        status: IDLE_STATUS,
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
        tableName: sanitizeTableName(name),
        id: guidGenerator(),
        preview: undefined,
        sizeInBytes: undefined,
        status: IDLE_STATUS,
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