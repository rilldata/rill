import {
  EntityType,
  StateType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  PersistentSourceEntity,
  PersistentSourceEntityService,
  PersistentSourceStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import type {
  DerivedSourceEntity,
  DerivedSourceEntityService,
  DerivedSourceStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/DerivedSourceEntityService";
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

export type EntityStateServicesMapType = {
  [EntityType.Source]?: {
    [StateType.Persistent]?: PersistentSourceEntityService;
    [StateType.Derived]?: DerivedSourceEntityService;
  };
  [EntityType.Model]?: {
    [StateType.Persistent]?: PersistentModelEntityService;
    [StateType.Derived]?: DerivedModelEntityService;
  };
  [EntityType.Application]?: {
    [StateType.Persistent]?: never;
    [StateType.Derived]?: ApplicationStateService;
  };
};

export type EntityRecordMapType = {
  [EntityType.Source]: {
    [StateType.Persistent]: PersistentSourceEntity;
    [StateType.Derived]: DerivedSourceEntity;
  };
  [EntityType.Model]: {
    [StateType.Persistent]: PersistentModelEntity;
    [StateType.Derived]: DerivedModelEntity;
  };
  [EntityType.Application]: {
    [StateType.Persistent]: never;
    [StateType.Derived]: ApplicationEntity;
  };
};
export type EntityStateActionArgMapType = {
  [EntityType.Source]: {
    [StateType.Persistent]: PersistentSourceStateActionArg;
    [StateType.Derived]: DerivedSourceStateActionArg;
  };
  [EntityType.Model]: {
    [StateType.Persistent]: PersistentModelStateActionArg;
    [StateType.Derived]: DerivedModelStateActionArg;
  };
  [EntityType.Application]: {
    [StateType.Persistent]: never;
    [StateType.Derived]: ApplicationStateActionArg;
  };
};
