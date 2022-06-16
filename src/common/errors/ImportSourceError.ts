import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export class ImportSourceError extends Error {
  public readonly errorType = ActionErrorType.ImportSource;
}
