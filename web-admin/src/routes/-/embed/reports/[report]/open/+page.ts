import { fetchResource } from "@rilldata/web-admin/features/resources/fetch-resource";
import {
  getCanvasReportOpenUrl,
  getOpenExploreName,
} from "@rilldata/web-admin/features/scheduled-reports/open/open-urls";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { isAIReportSpec } from "@rilldata/web-common/features/scheduled-reports/utils";
import { getRuntimeClient } from "@rilldata/web-common/runtime-client/v2/context";
import { error, redirect } from "@sveltejs/kit";

export async function load({ parent, url, params }) {
  // Embed routes don't have the project layout, so the runtime client is built from the embed layout's data.
  // The embed layout shows an error for missing params, so there's nothing to fetch.
  const { instanceId, runtimeHost, accessToken } = await parent();
  if (!instanceId || !runtimeHost || !accessToken) throw error(400);
  const client = getRuntimeClient({
    host: runtimeHost,
    instanceId,
    jwt: accessToken,
    authContext: "embed",
  });
  const report = await fetchResource(
    client,
    ResourceKind.Report,
    params.report,
  );
  const reportSpec = report?.report?.spec ?? {};

  // Canvas reports open the canvas dashboard directly, with the canvas state captured at scheduling time applied.
  const canvasUrl = getCanvasReportOpenUrl(reportSpec, {
    embedded: true,
  });
  if (canvasUrl) {
    redirect(307, canvasUrl);
  }

  return {
    reportId: params.report,
    reportSpec,
    // AI reports open an AI conversation, which embeds don't support.
    isAIReport: isAIReportSpec(reportSpec),
    executionTime: url.searchParams.get("execution_time"),
    exploreName: getOpenExploreName(reportSpec),
  };
}
