import { StateActions } from ".//StateActions";
import type {
  PersistentTableEntity,
  PersistentTableStateActionArg,
} from "$common/data-modeler-state-service/entity-state-service/PersistentTableEntityService";

export class TableStateActions extends StateActions {
  @StateActions.PersistentTableAction()
  public addOrUpdateTableToState(
    { stateService, draftState }: PersistentTableStateActionArg,
    table: PersistentTableEntity,
    isNew: boolean
  ): void {
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
}
