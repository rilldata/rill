import type { ValidationState } from "./MetricsDefinitionEntityService";
import type {
  EntityRecord,
  EntityState,
} from "./EntityStateService";
import {
  EntityStateActionArg,
  EntityStateService,
  EntityType,
  StateType,
} from "./EntityStateService";
import type { NicelyFormattedTypes } from "$web-local/lib/util/humanize-numbers";

export interface BasicMeasureDefinition {
  id: string;
  // mandatory user defined metadata
  expression: string;
  // optional user defined metadata
  sqlName?: string;
}
export interface MeasureDefinitionEntity
  extends EntityRecord,
    BasicMeasureDefinition {
  metricsDefId: string;
  creationTime: number;
  label?: string;
  description?: string;
  formatPreset?: NicelyFormattedTypes;
  expressionIsValid?: ValidationState;
  expressionValidationError?: string;
  sqlNameIsValid?: ValidationState;
}

export type MeasureDefinitionState = EntityState<MeasureDefinitionEntity>;

export type MeasureDefinitionStateActionArg = EntityStateActionArg<
  MeasureDefinitionEntity,
  MeasureDefinitionState,
  MeasureDefinitionStateService
>;

export class MeasureDefinitionStateService extends EntityStateService<
  MeasureDefinitionEntity,
  MeasureDefinitionState
> {
  public readonly entityType = EntityType.MeasureDefinition;
  public readonly stateType = StateType.Persistent;
}

// we need a fallback for measure name. this is needed when sqlName is not entered.
export function getFallbackMeasureName(index: number, sqlName?: string) {
  return sqlName?.length ? sqlName : `measure_${index}`;
}
