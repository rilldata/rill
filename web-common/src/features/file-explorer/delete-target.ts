import type { V1ResourceName } from "@rilldata/web-common/runtime-client";
import { ResourceKind } from "../entity-management/resource-selectors";
import {
  type Directory,
  findDirectory,
  flattenDirectory,
} from "./transform-file-list";

/**
 * The resource kinds that the delete confirmation names explicitly.
 * A file that defines any other kind of resource, or none, is described as a plain file.
 */
export type DeleteTargetKind =
  | "canvas"
  | "explore"
  | "metrics_view"
  | "model"
  | "source";

/** What a pending delete in the file explorer will remove; shown to the user before it runs. */
export type DeleteTarget =
  | {
      type: "file";
      path: string;
      /** The resource the file defines, when the runtime has parsed one of a named kind. */
      resource?: { kind: DeleteTargetKind; name: string };
    }
  | {
      type: "folder";
      path: string;
      /** Paths of every file in the folder, including files in nested folders. */
      files: string[];
    };

const NAMED_KINDS: Partial<Record<ResourceKind, DeleteTargetKind>> = {
  [ResourceKind.Canvas]: "canvas",
  [ResourceKind.Explore]: "explore",
  [ResourceKind.MetricsView]: "metrics_view",
  [ResourceKind.Model]: "model",
  [ResourceKind.Source]: "source",
};

export function getFileDeleteTarget(
  path: string,
  resourceName: V1ResourceName | undefined,
): DeleteTarget {
  const kind = NAMED_KINDS[resourceName?.kind as ResourceKind];
  if (!kind || !resourceName?.name) return { type: "file", path };
  return { type: "file", path, resource: { kind, name: resourceName.name } };
}

/**
 * Returns the delete target for a folder, or undefined when the folder is empty
 * (or not in the tree) and can be deleted without asking.
 */
export function getFolderDeleteTarget(
  tree: Directory,
  path: string,
): DeleteTarget | undefined {
  const dir = findDirectory(tree, path);
  if (!dir || (!dir.directories.length && !dir.files.length)) return undefined;
  return { type: "folder", path, files: flattenDirectory(dir) };
}
