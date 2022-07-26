import type { DerivedModelState } from "$common/data-modeler-state-service/entity-state-service/DerivedModelEntityService";

export const selectDerivedModelById = (
  derivedModelState: DerivedModelState,
  id: string
) => {
  return derivedModelState.entities.find((model) => model.id === id);
};
