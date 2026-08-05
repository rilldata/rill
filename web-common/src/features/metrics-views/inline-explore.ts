import { goto } from "$app/navigation";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { YAMLMap, parseDocument } from "yaml";
import {
  getFileHref,
  navigateToExplore,
  withEditorPrefix,
} from "../../layout/navigation/editor-routing";
import { previewModeStore } from "../../layout/preview-mode-store";
import type { WorkspaceView } from "../../layout/workspace/workspace-stores";
import { waitUntil } from "../../lib/waitUtils";
import { fileArtifacts } from "../entity-management/file-artifacts";

// Views supported by the metrics view workspace: the base views plus the inline
// explore dashboard editor.
export type MetricsWorkspaceView = WorkspaceView | "explore";

export type InlineExploreState = {
  // True when the file opts out of the legacy behavior where an explore is
  // auto-emitted for every metrics view (i.e. it sets `version` > 0).
  isLatestVersion: boolean;
  // True when the file defines an inline explore: a `explore:` mapping that is
  // not `skip: true`. Only meaningful for latest-version metrics views.
  exploreEnabled: boolean;
  // The inline explore's `name` override, or null to default to the metrics view name.
  exploreName: string | null;
};

// parseInlineExploreState inspects a metrics view file's YAML for the inline
// explore configuration. It never throws: unparseable content yields all-false state.
export function parseInlineExploreState(
  content: string | null | undefined,
): InlineExploreState {
  const doc = parseDocument(content ?? "");

  const isLatestVersion = Number(doc.get("version")) > 0;

  const explore = doc.get("explore");
  const exploreEnabled =
    isLatestVersion &&
    explore instanceof YAMLMap &&
    explore.get("skip") !== true;

  const rawName = explore instanceof YAMLMap ? explore.get("name") : undefined;
  const exploreName =
    typeof rawName === "string" && rawName !== "" ? rawName : null;

  return { isLatestVersion, exploreEnabled, exploreName };
}

// previewInlineExplore opens the explore emitted by a metrics view file's inline
// explore block: the explore page in preview mode, or the metrics view file's
// explore view in the editor. It routes based on the file's actual YAML rather
// than assuming an inline explore exists (e.g. AI generation is not guaranteed to
// emit one): without an inline explore it falls back to the metrics view itself.
export async function previewInlineExplore(
  queryClient: QueryClient,
  filePath: string,
) {
  const artifact = fileArtifacts.getFileArtifact(filePath);
  const resource = artifact.getResource(queryClient);
  await waitUntil(() => get(resource).data !== undefined, 10000);

  const metricsViewName = get(resource).data?.meta?.name?.name;
  if (!metricsViewName) {
    throw new Error(`Failed to resolve the metrics view for ${filePath}`);
  }

  const { isLatestVersion, exploreEnabled, exploreName } =
    parseInlineExploreState(
      get(artifact.editorContent) ?? get(artifact.remoteContent),
    );

  if (get(previewModeStore)) {
    // Legacy metrics views auto-emit an explore named after the metrics view.
    if (exploreEnabled || !isLatestVersion) {
      await navigateToExplore(exploreName ?? metricsViewName);
    } else {
      await goto(withEditorPrefix("/dashboards"));
    }
    return;
  }

  await goto(getFileHref(filePath, exploreEnabled ? "explore" : undefined));
}

// enableInlineExploreAndPreview ensures the metrics view file defines an inline
// explore, then opens it. Legacy (unversioned) metrics views already auto-emit an
// explore named after the metrics view, so those only navigate.
export async function enableInlineExploreAndPreview(
  queryClient: QueryClient,
  filePath: string,
) {
  const artifact = fileArtifacts.getFileArtifact(filePath);
  const content =
    get(artifact.editorContent) ?? get(artifact.remoteContent) ?? "";
  const { isLatestVersion, exploreEnabled } = parseInlineExploreState(content);

  if (!isLatestVersion) {
    const resource = artifact.getResource(queryClient);
    await waitUntil(() => get(resource).data !== undefined, 10000);
    const metricsViewName = get(resource).data?.meta?.name?.name;
    if (!metricsViewName) {
      throw new Error(`Failed to resolve the metrics view for ${filePath}`);
    }
    await navigateToExplore(metricsViewName);
    return;
  }

  if (!exploreEnabled) {
    const doc = parseDocument(content);
    const explore = doc.get("explore");
    if (explore instanceof YAMLMap) {
      explore.delete("skip");
    } else {
      doc.set("explore", {});
    }
    artifact.updateEditorContent(doc.toString(), false, true);
  }

  await previewInlineExplore(queryClient, filePath);
}
