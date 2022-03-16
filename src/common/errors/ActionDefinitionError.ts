import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export class ActionDefinitionError extends Error {
    public readonly errorType = ActionErrorType.ActionDefinition;
}
