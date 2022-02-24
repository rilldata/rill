import {
    EntityType,
    StateType
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
    PersistentTableEntity,
    PersistentTableEntityService, PersistentTableStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type {
    DerivedTableEntity,
    DerivedTableEntityService, DerivedTableStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type {
    PersistentModelEntity,
    PersistentModelEntityService, PersistentModelStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
    DerivedModelEntity,
    DerivedModelEntityService, DerivedModelStateActionArg
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";

export type EntityStateServicesMapType = {
    [EntityType.Table]?: {
        [StateType.Persistent]?: PersistentTableEntityService,
        [StateType.Derived]?: DerivedTableEntityService,
    },
    [EntityType.Model]?: {
        [StateType.Persistent]?: PersistentModelEntityService,
        [StateType.Derived]?: DerivedModelEntityService,
    },
};

export type EntityRecordMapType = {
    [EntityType.Table]: {
        [StateType.Persistent]: PersistentTableEntity,
        [StateType.Derived]: DerivedTableEntity,
    },
    [EntityType.Model]: {
        [StateType.Persistent]: PersistentModelEntity,
        [StateType.Derived]: DerivedModelEntity,
    },
};
export type EntityStateActionArgMapType = {
    [EntityType.Table]: {
        [StateType.Persistent]: PersistentTableStateActionArg,
        [StateType.Derived]: DerivedTableStateActionArg,
    },
    [EntityType.Model]: {
        [StateType.Persistent]: PersistentModelStateActionArg,
        [StateType.Derived]: DerivedModelStateActionArg,
    },
};
