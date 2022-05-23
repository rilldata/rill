import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export class ImportTableError extends Error {
  public readonly errorType = ActionErrorType.ImportTable;
}
