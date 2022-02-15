import type {DataModelerState, Table, Model} from "$lib/types";
import {guidGenerator} from "$lib/util/guid";
import {sanitizeQuery as _sanitizeQuery} from "$lib/util/sanitize-query";
import {IDLE_STATUS} from "$common/constants";
import {sanitizeTableName} from "$lib/util/sanitize-table-name";

let modelNumber = 0;

interface NewModelArguments {
    query?: string;
    name?: string;
}

export function getNewTable(): Table {
    return {
        id: guidGenerator(),
        path: '',
        tableName: '',
        profile: [],
        cardinality: undefined,
        sizeInBytes: undefined,
        head: [],
        status: IDLE_STATUS,
        lastUpdated: 0,
    }
}

export function getNewModel(params: NewModelArguments = {}): Model {
    const query = params.query || '';
    const sanitizedQuery = _sanitizeQuery(query);
    const name = `${params.name || `query_${modelNumber}`}.sql`;
    modelNumber += 1;
    return {
        query,
        sanitizedQuery,
        name,
        tableName: sanitizeTableName(name),
        id: guidGenerator(),
        preview: undefined,
        sizeInBytes: undefined,
        status: IDLE_STATUS,
        sources: [],
    };
}

export function getEmptyModel(): Model {
    return getNewModel({});
}

export function initialState() : DataModelerState {
    return {
        models: [getEmptyModel()],
        tables: [],
        metricsModels: [],
        exploreConfigurations: [],
        status: 'disconnected'
    }
}