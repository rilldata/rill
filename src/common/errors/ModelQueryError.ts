import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export class ModelQueryError extends Error {
    public readonly errorType = ActionErrorType.ModelQuery;
}
