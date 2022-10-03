import type { PersistentTableStateActionArg } from "./entity-state-service/PersistentTableEntityService";
import { StateActions } from ".//StateActions";
import type { DerivedTableStateActionArg } from "./entity-state-service/DerivedTableEntityService";

export class TableStateActions extends StateActions {
  @StateActions.PersistentTableAction()
  public addOrUpdateTableToState(): void {
    // we do not create any state here since it is contingent on whether the import step
    // worked as expected.
  }

  @StateActions.PersistentTableAction()
  public updateTableName(
    { stateService, draftState }: PersistentTableStateActionArg,
    tableId: string,
    name: string
  ): void {
    // update both "name" and "tableName" while both properties exist in the data model
    stateService.updateEntityField(draftState, tableId, "name", name);
    stateService.updateEntityField(draftState, tableId, "tableName", name);
  }

  @StateActions.DerivedTableAction()
  public updateTablePreview(
    { stateService, draftState }: DerivedTableStateActionArg,
    tableId: string,
    preview: Array<unknown>
  ): void {
    stateService.updateEntityField(draftState, tableId, "preview", preview);
  }
}
