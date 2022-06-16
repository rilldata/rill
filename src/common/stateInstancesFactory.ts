import type { DataModelerState, Model } from "$lib/types";
import { guidGenerator } from "$lib/util/guid";
import {
  extractSourceName,
  sanitizeSourceName,
} from "$lib/util/extract-source-name";
import type { PersistentSourceEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentSourceEntityService";
import {
  EntityStatus,
  EntityType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type { DerivedSourceEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedSourceEntityService";

interface NewModelArguments {
  query?: string;
  name?: string;
}

export function getNewSource(): PersistentSourceEntity {
  return {
    id: guidGenerator(),
    type: EntityType.Source,
    path: "",
    lastUpdated: 0,
  };
}
export function getNewDerivedSource(
  source: PersistentSourceEntity
): DerivedSourceEntity {
  return {
    id: source.id,
    type: EntityType.Source,
    profile: [],
    lastUpdated: 0,
    status: EntityStatus.Idle,
  };
}

export function cleanModelName(name: string): string {
  return name.replace(/\.sql$/, "");
}
export function getNewModel(
  params: NewModelArguments = {},
  modelNumber
): PersistentModelEntity {
  const query = params.query || "";
  const name = `${
    params.name ? cleanModelName(params.name) : `query_${modelNumber}`
  }.sql`;
  return {
    id: guidGenerator(),
    type: EntityType.Model,
    query,
    name,
    sourceName: sanitizeSourceName(extractSourceName(name)),
    lastUpdated: 0,
  };
}
export function getNewDerivedModel(
  model: PersistentModelEntity
): DerivedModelEntity {
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

export function initialState(): DataModelerState {
  return {
    models: [],
    sources: [],
    metricsModels: [],
    exploreConfigurations: [],
    status: "disconnected",
  };
}
