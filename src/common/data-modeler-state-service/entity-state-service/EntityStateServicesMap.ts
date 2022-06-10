import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  PersistentTableEntity,
  PersistentTableEntityService,
  PersistentTableStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import type {
  DerivedTableEntity,
  DerivedTableEntityService,
  DerivedTableStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type {
  PersistentModelEntity,
  PersistentModelEntityService,
  PersistentModelStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type {
  DerivedModelEntity,
  DerivedModelEntityService,
  DerivedModelStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type {
  ApplicationEntity,
  ApplicationStateActionArg,
  ApplicationStateService,
} from "$common/data-modeler-state-service/entity-state-service/ApplicationEntityService";
import type {
  MetricsDefinitionEntity,
  MetricsDefinitionStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { MetricsDefinitionStateService } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export type EntityStateServicesMapType = {
  [EntityType.Table]?: {
    [StateType.Persistent]?: PersistentTableEntityService;
    [StateType.Derived]?: DerivedTableEntityService;
  };
  [EntityType.Model]?: {
    [StateType.Persistent]?: PersistentModelEntityService;
    [StateType.Derived]?: DerivedModelEntityService;
  };
  [EntityType.Application]?: {
    [StateType.Persistent]?: never;
    [StateType.Derived]?: ApplicationStateService;
  };
  [EntityType.MetricsDefinition]?: {
    [StateType.Persistent]?: MetricsDefinitionStateService;
    [StateType.Derived]?: never;
  };
};

export type EntityRecordMapType = {
  [EntityType.Table]: {
    [StateType.Persistent]: PersistentTableEntity;
    [StateType.Derived]: DerivedTableEntity;
  };
  [EntityType.Model]: {
    [StateType.Persistent]: PersistentModelEntity;
    [StateType.Derived]: DerivedModelEntity;
  };
  [EntityType.Application]: {
    [StateType.Persistent]: never;
    [StateType.Derived]: ApplicationEntity;
  };
  [EntityType.MetricsDefinition]: {
    [StateType.Persistent]: MetricsDefinitionEntity;
    [StateType.Derived]: never;
  };
};
export type EntityStateActionArgMapType = {
  [EntityType.Table]: {
    [StateType.Persistent]: PersistentTableStateActionArg;
    [StateType.Derived]: DerivedTableStateActionArg;
  };
  [EntityType.Model]: {
    [StateType.Persistent]: PersistentModelStateActionArg;
    [StateType.Derived]: DerivedModelStateActionArg;
  };
  [EntityType.Application]: {
    [StateType.Persistent]: never;
    [StateType.Derived]: ApplicationStateActionArg;
  };
  [EntityType.MetricsDefinition]: {
    [StateType.Persistent]: MetricsDefinitionStateActionArg;
    [StateType.Derived]: never;
  };
};
