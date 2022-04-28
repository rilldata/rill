import type { DataModelerState, Model } from "$lib/types";
import { guidGenerator } from "$lib/util/guid";
import { extractTableName, sanitizeTableName } from "$lib/util/extract-table-name";
import type {
    PersistentTableEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import { EntityStatus, EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    PersistentModelEntity
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
    DerivedModelEntity
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import { sanitizeQuery } from "$lib/util/sanitize-query";
import type {
    DerivedTableEntity
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";

let modelNumber = 0;

interface NewModelArguments {
    query?: string;
    name?: string;
}

export function getNewTable(): PersistentTableEntity {
    return {
        id: guidGenerator(),
        type: EntityType.Table,
        path: '',
        lastUpdated: 0,
    };
}
export function getNewDerivedTable(table: PersistentTableEntity): DerivedTableEntity {
    return {
        id: table.id,
        type: EntityType.Table,
        profile: [],
        lastUpdated: 0,
        status: EntityStatus.Idle,
    };
}

export function cleanModelName(name: string): string {
    return name.replace(/\.sql$/, "");
}
export function getNewModel(params: NewModelArguments = {}): PersistentModelEntity {
    const query = params.query || '';
    const name = `${params.name ? cleanModelName(params.name) : `query_${modelNumber}`}.sql`;
    modelNumber += 1;
    return {
        id: guidGenerator(),
        type: EntityType.Model,
        query,
        name,
        tableName: sanitizeTableName(extractTableName(name)),
        lastUpdated: 0,
    };
}
export function getNewDerivedModel(model: PersistentModelEntity): DerivedModelEntity {
    return {
        id: model.id,
        type: EntityType.Model,
        // do not assign this to trigger profiling
        sanitizedQuery: "",
        profile: [],
        lastUpdated: 0,
        status: EntityStatus.Idle,
    };
}

export function getEmptyModel(): Model {
    return getNewModel({}) as any;
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
