import { ActionErrorType } from "../data-modeler-service/response/ActionResponseMessage";
import type { TypedError } from "./TypedError";

export class ActionDefinitionError extends Error implements TypedError {
  public readonly errorType = ActionErrorType.ActionDefinition;
}
