import { StateActions } from ".//StateActions";

export class TableStateActions extends StateActions {
  @StateActions.PersistentTableAction()
  public addOrUpdateTableToState(): void {
    // we do not create any state here since it is contingent on whether the import step
    // worked as expected.
  }
}
