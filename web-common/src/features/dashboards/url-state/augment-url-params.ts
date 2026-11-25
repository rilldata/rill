import { page } from "$app/state";
import { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { convertProtoToUrlParams } from "@rilldata/web-common/features/dashboards/proto-state/convert-proto-to-url-params.ts";

/**
 * Builds the `web_open_state` of an alert or a report:
 * the existing state (the saved one when editing, else the current page's url) with the form's filters applied.
 *
 * Explore state is stored with a leading `?` to tell it apart from the legacy proto state.
 * Canvas state was always url params, so it is stored as is.
 */
export function augmentUrlParams(
  kind: "explore" | "canvas",
  existingAnnotations: Record<string, string> | undefined,
  expressionFilterManager: ExpressionFilterManager,
  timeFilterManager: TimeFilterManager,
  addParams?: (urlParams: URLSearchParams) => void,
) {
  const urlParams = existingAnnotations
    ? convertStateParamToUrlParams(kind, existingAnnotations["web_open_state"])
    : new URLSearchParams(page.url.search);

  expressionFilterManager.applyFilterToParams(urlParams);
  timeFilterManager.applyFilterToParams(urlParams);

  addParams?.(urlParams);

  return kind === "explore" ? "?" + urlParams.toString() : urlParams.toString();
}

/**
 * Parses a saved `web_open_state`.
 * Explore state without a leading `?` predates url params and is a base64 encoded proto.
 */
export function convertStateParamToUrlParams(
  kind: "explore" | "canvas",
  state: string | undefined,
) {
  if (!state) return new URLSearchParams();
  if (kind === "canvas" || state.startsWith("?")) {
    return new URLSearchParams(state);
  }
  return convertProtoToUrlParams(state);
}
