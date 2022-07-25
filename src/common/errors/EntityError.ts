import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";
import type { TypedError } from "$common/errors/TypedError";

export class EntityError extends Error implements TypedError {
  public readonly errorType = ActionErrorType.EntityError;
}
