import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export class EntityError extends Error {
  public readonly errorType = ActionErrorType.EntityError;
}
