import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store.ts";
import { removeEmbedParams } from "@rilldata/web-admin/features/embeds/init-embed-public-api.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { redirect } from "@sveltejs/kit";

export const load = async ({ url }) => {
  const embedStore = EmbedStore.getInstance();
  if (!embedStore) {
    EmbedStore.init(url);

    // Open links from report and alert emails open the report or alert instead of a dashboard.
    // GetIFrame decodes and validates the link's `rill_open` param, and passes the resource to open as plain params.
    // `open_report` is the param of earlier versions of GetIFrame, which only supported reports.
    // It's still accepted, since the admin service and the frontend may be deployed at different times.
    const legacyOpenReport = url.searchParams.get("open_report");
    const openKind =
      url.searchParams.get("open_kind") ?? (legacyOpenReport ? "report" : null);
    const openName = url.searchParams.get("open_name") ?? legacyOpenReport;
    const openSection =
      openKind === "report" ? "reports" : openKind === "alert" ? "alerts" : "";
    if (openSection && openName) {
      const path = `/-/embed/${openSection}/${encodeURIComponent(openName)}/open`;
      const executionTime = url.searchParams.get("open_execution_time");
      throw redirect(
        307,
        executionTime
          ? `${path}?${new URLSearchParams({ execution_time: executionTime }).toString()}`
          : path,
      );
    }

    const resource = url.searchParams.get("resource");
    if (!resource) {
      throw redirect(307, "/-/embed");
    }

    const type =
      url.searchParams.get("type") === ResourceKind.Canvas
        ? "canvas"
        : "explore";
    // Retain non-embed search params
    const nonEmbedSearchParams = removeEmbedParams(url.searchParams);
    throw redirect(
      307,
      `/-/embed/${type}/${resource}?${nonEmbedSearchParams.toString()}`,
    );
  }

  const {
    instanceId,
    runtimeHost,
    accessToken,
    missingRequireParams,
    navigationEnabled,
    navigationBarEnabled,
    visibleExplores,
    adminToken,
  } = embedStore;

  return {
    instanceId,
    runtimeHost,
    accessToken,
    missingRequireParams,
    navigationEnabled,
    navigationBarEnabled,
    visibleExplores,
    // The admin token is only issued for embeds with an external user ID, which identifies the owner of reports created from the embed.
    embedAdminActions: !!adminToken,
  };
};
