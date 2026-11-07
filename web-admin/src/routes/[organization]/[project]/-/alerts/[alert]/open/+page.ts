import { getCloudRuntimeClient } from "@rilldata/web-admin/lib/runtime-client";
import { getExploreUrlForMetricsSQL } from "@rilldata/web-common/features/explore-mappers/open-metrics-sql";
import { getExploreName } from "@rilldata/web-common/features/explore-mappers/utils";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  const { alert, runtime } = await parent();
  const organization = params.organization;
  const project = params.project;
  const alertId = params.alert;
  const executionTime = url.searchParams.get("execution_time");
  const token = url.searchParams.get("token");
  const exploreName =
    alert.alert.spec.annotations["explore"] ??
    getExploreName(alert.alert.spec.annotations?.web_open_path); // backwards compatibility

  // An alert written in metrics SQL opens the explore its compiled query maps to.
  // When it can't be shown faithfully, the alert page is the honest fallback.
  const spec = alert.alert.spec;
  if (spec?.resolver === "metrics_sql") {
    let exploreURL: string;
    try {
      exploreURL = await getExploreUrlForMetricsSQL({
        client: getCloudRuntimeClient(runtime),
        sql: spec.resolverProperties?.sql as string,
        exploreName,
        executionTime,
        token,
        organization,
        project,
      });
    } catch (e) {
      console.error("Failed to open the alert's explore:", e);
      exploreURL = `/${organization}/${project}/-/alerts/${alertId}`;
    }
    redirect(307, exploreURL);
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
