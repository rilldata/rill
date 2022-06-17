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
}
