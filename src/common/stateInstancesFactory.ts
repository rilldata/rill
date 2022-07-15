import type { DataModelerState } from "$lib/types";
import { guidGenerator } from "$lib/util/guid";
import {
  extractTableName,
  sanitizeTableName,
} from "$lib/util/extract-table-name";
import type { PersistentTableEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";
import {
  EntityStatus,
  EntityType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { PersistentModelEntity } from "$common/data-modeler-state-service/entity-state-service/PersistentModelEntityService";
import type { DerivedModelEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";
import type { DerivedTableEntity } from "$common/data-modeler-state-service/entity-state-service/DerivedTableEntityService";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";

interface NewModelArguments {
  query?: string;
  name?: string;
}

export function getNewTable(): PersistentTableEntity {
  return {
    id: guidGenerator(),
    type: EntityType.Table,
    path: "",
    lastUpdated: 0,
  };
}
export function getNewDerivedTable(
  table: PersistentTableEntity
): DerivedTableEntity {
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
    tableName: sanitizeTableName(extractTableName(name)),
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

export function getMetricsDefinition(counter: number): MetricsDefinitionEntity {
  return {
    id: guidGenerator(),
    type: EntityType.MetricsDefinition,
    creationTime: Date.now(),
    metricDefLabel: `metrics ${counter}`,
    sourceModelId: undefined,
    timeDimension: undefined,
    measureIds: [],
    dimensionIds: [],
    lastUpdated: 0,
  };
}

export function getMeasureDefinition(
  metricsDefId: string,
  expression = ""
): MeasureDefinitionEntity {
  return {
    id: guidGenerator(),
    creationTime: Date.now(),
    metricsDefId,
    type: EntityType.MeasureDefinition,
    expression,
    lastUpdated: 0,
  };
}

export function getDimensionDefinition(
  metricsDefId: string
): DimensionDefinitionEntity {
  return {
    id: guidGenerator(),
    creationTime: Date.now(),
    metricsDefId,
    type: EntityType.DimensionDefinition,
    dimensionColumn: "",
    lastUpdated: 0,
  };
}

export function initialState(): DataModelerState {
  return {
    models: [],
    tables: [],
    metricsModels: [],
    exploreConfigurations: [],
    status: "disconnected",
  };
}
