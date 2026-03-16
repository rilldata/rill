export const ssr = false;

import { redirect } from "@sveltejs/kit";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
  type V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client/index.js";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.js";
import { handleUninitializedProject } from "@rilldata/web-common/features/welcome/is-project-initialized.js";
import { localServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { getLocalRuntimeClient } from "../lib/runtime-client";
import {
  DEVELOPER_ALLOWED_PREFIXES,
  PREVIEW_ALLOWED_PREFIXES,
} from "./route-constants";
import { Settings } from "luxon";

Settings.defaultLocale = "en";

// Cache metadata: previewMode is static for the server lifetime, so fetch once
let cachedMetadata: Awaited<ReturnType<typeof localServiceGetMetadata>> | null =
  null;

export async function load({ url, depends, untrack }) {
  depends("app:init");

  // Fetch metadata to check preview mode (cached after first load)
  if (!cachedMetadata) {
    cachedMetadata = await localServiceGetMetadata();
  }
  const metadata = cachedMetadata;
  const previewMode = true; // TEMP: force preview mode for testing

  // Enforce mode-based route locking
  if (previewMode) {
    // Preview mode: only allow preview-related and shared routes
    const isAllowed = PREVIEW_ALLOWED_PREFIXES.some((prefix) =>
      url.pathname.startsWith(prefix),
    );
    if (!isAllowed) {
      eventBus.emit("notification", {
        message: "This page is only available in Developer mode",
      });
      throw redirect(303, "/dashboards");
    }
  } else {
    // Developer mode: block preview-exclusive routes
    const isAllowed =
      url.pathname === "/" ||
      DEVELOPER_ALLOWED_PREFIXES.some((prefix) =>
        url.pathname.startsWith(prefix),
      );
    if (!isAllowed) {
      eventBus.emit("notification", {
        message: "This page is only available in Preview mode",
      });
      throw redirect(303, "/");
    }
  }

  const client = getLocalRuntimeClient();

  // Set the client on fileArtifacts early so child page load functions
  // (e.g., files/[...file]/+page.ts) can access it before components render.
  fileArtifacts.setClient(client);

  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(client.instanceId, {}),
    queryFn: ({ signal }) => {
      return runtimeServiceListFiles(client, {}, { signal });
    },
  });

  const firstDashboardFile = files.files?.find((file) =>
    file.path?.startsWith("/dashboards/"),
  );

  let initialized = !!files.files?.some(({ path }) => path === "/rill.yaml");

  const redirectPath = untrack(() => {
    if (!url.searchParams.get("redirect")) return false;

    // In preview mode, redirect to /dashboards instead of /files
    if (previewMode) {
      return url.pathname !== "/dashboards" && "/dashboards";
    }

    return (
      url.pathname !== `/files${firstDashboardFile?.path}` &&
      `/files${firstDashboardFile?.path}`
    );
  });

  if (!initialized) {
    initialized = await handleUninitializedProject(client);
  } else if (redirectPath) {
    throw redirect(303, redirectPath);
  }

  return { initialized, previewMode, metadata };
}
