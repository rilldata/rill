import { ActionErrorType } from "$common/data-modeler-service/response/ActionResponseMessage";

export class QueryCancelledError extends Error {
    public readonly errorType = ActionErrorType.QueryCancelled;

    constructor() {
        super("Query cancelled");
    }
}
