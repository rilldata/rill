import { ActionErrorType } from "../data-modeler-service/response/ActionResponseMessage";
import type { TypedError } from "./TypedError";

export class EntityError extends Error implements TypedError {
  public readonly errorType = ActionErrorType.EntityError;
}
