import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import { isAIAlertSpec } from "@rilldata/web-common/features/alerts/ai-alert-utils";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  const { alert } = await parent();
  const organization = params.organization;
  const project = params.project;
  const alertId = params.alert;
  const executionTime = url.searchParams.get("execution_time");
  const token = url.searchParams.get("token");
  const exploreName =
    alert.alert.spec.annotations["explore"] ??
    getExploreName(alert.alert.spec.annotations?.web_open_path); // backwards compatibility

  // AI alerts open the AI conversation behind the check that triggered.
  // The runtime adds the session ID to the open link, which keeps opening the conversation even if the alert was switched to criteria since.
  // In creator mode the link also carries a magic token, which grants read access to the shared conversation and must be forwarded.
  const sessionId = url.searchParams.get("session_id");
  if (sessionId || isAIAlertSpec(alert.alert.spec)) {
    const path = sessionId
      ? `/${organization}/${project}/-/ai/${sessionId}`
      : `/${organization}/${project}/-/ai`;
    const search = token ? new URLSearchParams({ token }).toString() : "";
    redirect(307, search ? `${path}?${search}` : path);
  }

  return {
    organization,
    project,
    alertId,
    alert,
    executionTime,
    token,
    exploreName,
  };
}
