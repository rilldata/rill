import { getExploreName } from "@rilldata/web-admin/features/dashboards/query-mappers/utils";
import type { V1ReportSpec } from "@rilldata/web-common/runtime-client";

export function getDashboardNameFromReport(
  reportSpec: V1ReportSpec | undefined,
): string | null {
  if (!reportSpec?.queryArgsJson) return null;
  if (reportSpec.annotations?.web_open_path)
    return getExploreName(reportSpec.annotations.web_open_path);

  const queryArgsJson = JSON.parse(reportSpec.queryArgsJson);
  return (
    queryArgsJson?.metrics_view_name ??
    queryArgsJson?.metricsViewName ??
    queryArgsJson?.metrics_view ??
    queryArgsJson?.metricsView ??
    null
  );
}
