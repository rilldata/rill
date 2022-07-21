import type { ActionResponse } from "$common/data-modeler-service/response/ActionResponse";
import { ActionStatus } from "$common/data-modeler-service/response/ActionResponse";
import notifications from "$lib/components/notifications";

export function handleErrorResponse(actionResponse: ActionResponse) {
  if (!actionResponse || actionResponse.status === ActionStatus.Success) return;

  const actionResponseMessage = actionResponse.messages[0];
  notifications.send({ message: actionResponseMessage.message });
}
