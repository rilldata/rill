import { dataModelerStateService } from "../../application-state-stores/application-store";
import {
  EntityType,
  StateType,
} from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";

export const selectPersistentModelById = (id: string) =>
  dataModelerStateService
    .getEntityStateService(EntityType.Model, StateType.Persistent)
    .getById(id);

export const selectDerivedModelById = (id: string) =>
  dataModelerStateService
    .getEntityStateService(EntityType.Model, StateType.Derived)
    .getById(id);

export const selectDerivedModelBySourceName = (persistentTableName: string) =>
  dataModelerStateService
    .getEntityStateService(EntityType.Model, StateType.Derived)
    .getCurrentState()
    .entities.filter(
      (derivedModel) =>
        !!derivedModel.sources?.find(
          (sourceTable) => sourceTable.name === persistentTableName
        )
    );
