import { getOpenExploreName } from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
import { isAIAlertSpec } from "@rilldata/web-common/features/alerts/ai-alert-utils";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  const { alert } = await parent();
  const alertSpec = alert?.alert?.spec ?? {};
  const token = url.searchParams.get("token");

  // AI alerts open the AI conversation behind the check that fired, as AI reports do.
  // The runtime adds the session ID to the open link, which keeps opening the conversation if the alert was switched to criteria since.
  // In creator mode the link also carries a magic token, which grants read access to the shared conversation and must be forwarded.
  const sessionId = url.searchParams.get("session_id");
  if (sessionId || isAIAlertSpec(alertSpec)) {
    const path = sessionId
      ? `/${params.organization}/${params.project}/-/ai/${sessionId}`
      : `/${params.organization}/${params.project}/-/ai`;
    const search = token ? new URLSearchParams({ token }).toString() : "";
    redirect(307, search ? `${path}?${search}` : path);
  }

  return {
    organization: params.organization,
    project: params.project,
    alertId: params.alert,
    alertSpec,
    executionTime: url.searchParams.get("execution_time"),
    token,
    exploreName: getOpenExploreName(alertSpec),
  };
}
