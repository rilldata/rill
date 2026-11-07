import { getPdfOptionsFromWebOpenState } from "@rilldata/web-common/features/scheduled-reports/utils";
import { V1ExportFormat } from "@rilldata/web-common/runtime-client";
import { redirect } from "@sveltejs/kit";

export async function load({ parent, url }) {
  const { report } = await parent();
  const spec = report?.report?.spec;

  const canvasName = spec?.annotations?.canvas;
  const isCanvasPdf =
    spec?.exportFormat === V1ExportFormat.EXPORT_FORMAT_PDF && !!canvasName;
  if (!isCanvasPdf) {
    return { canvasName: undefined };
  }

  // Canvas state is URL-param based. The report stores the canvas state captured at
  // scheduling time in the web_open_state annotation; merge it into the URL (preserving
  // token and execution_time) so the canvas renders with that state applied.
  const stateParams = new URLSearchParams(
    spec.annotations?.web_open_state ?? "",
  );
  const missingStateParam = [...stateParams.keys()].some(
    (key) => !url.searchParams.has(key),
  );
  if (missingStateParam) {
    const merged = new URL(url);
    stateParams.forEach((value, key) => {
      if (!merged.searchParams.has(key)) merged.searchParams.set(key, value);
    });
    redirect(307, `${merged.pathname}${merged.search}`);
  }

  return {
    canvasName,
    ...getPdfOptionsFromWebOpenState(spec.annotations?.web_open_state),
  };
}
