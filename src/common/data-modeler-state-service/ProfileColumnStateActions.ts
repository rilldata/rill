import { StateActions } from "./StateActions";
import type { ColumnarTypeKeys, ProfileColumnSummary } from "$lib/types";
import type { DataProfileStateActionArg } from "$common/data-modeler-state-service/entity-state-service/DataProfileEntity";
import type { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { shallowCopy } from "$common/utils/shallowCopy";

export enum ColumnarItemType {
  Table,
  Model,
}
export const ColumnarItemTypeMap: {
  [type in ColumnarItemType]: ColumnarTypeKeys;
} = {
  [ColumnarItemType.Table]: "tables",
  [ColumnarItemType.Model]: "models",
};

export class ProfileColumnStateActions extends StateActions {
  @StateActions.DerivedAction()
  public clearProfileSummary(
    { stateService, draftState }: DataProfileStateActionArg,
    entityType: EntityType,
    entityId: string
  ): void {
    const entityToUpdate = stateService.getById(entityId, draftState);
    entityToUpdate.profile?.forEach((profile) => {
      profile.summary = null;
      profile.nullCount = null;
    });
    // TODO: update this automatically
    entityToUpdate.lastUpdated = Date.now();
  }

  @StateActions.DerivedAction()
  public updateColumnSummary(
    { stateService, draftState }: DataProfileStateActionArg,
    entityType: EntityType,
    entityId: string,
    columnName: string,
    summary: ProfileColumnSummary
  ): void {
    const entityToUpdate = stateService.getById(entityId, draftState);
    const profileToUpdate = entityToUpdate.profile.find(
      (column) => column.name === columnName
    );
    profileToUpdate.summary ??= {};
    shallowCopy(summary, profileToUpdate.summary);
    // TODO: update this automatically
    entityToUpdate.lastUpdated = Date.now();
  }

  @StateActions.DerivedAction()
  public updateNullCount(
    { stateService, draftState }: DataProfileStateActionArg,
    entityType: EntityType,
    entityId: string,
    columnName: string,
    nullCount: number
  ): void {
    const entityToUpdate = stateService.getById(entityId, draftState);
    const profileToUpdate = entityToUpdate.profile.find(
      (column) => column.name === columnName
    );
    profileToUpdate.nullCount = nullCount;
    // TODO: update this automatically
    entityToUpdate.lastUpdated = Date.now();
  }

  @StateActions.DerivedAction()
  public markAsProfiled(
    { stateService, draftState }: DataProfileStateActionArg,
    entityType: EntityType,
    entityId: string,
    profiled: boolean
  ) {
    stateService.updateEntityField(draftState, entityId, "profiled", profiled);
  }
}
