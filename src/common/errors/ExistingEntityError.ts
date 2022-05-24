import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export class ExistingEntityError extends Error {
  public readonly errorType = ActionErrorType.ExistingEntityError;
}
