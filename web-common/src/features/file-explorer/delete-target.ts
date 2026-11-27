import type {
  V1Resource,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";
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
      /** Files of the resources that depend on the file's resource, directly or through others; they break once it is gone. */
      dependents: string[];
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

/**
 * Returns the delete target for a file.
 * resources is every resource in the project, used to find the files that depend on the file's resource.
 */
export function getFileDeleteTarget(
  path: string,
  resourceName: V1ResourceName | undefined,
  resources: V1Resource[],
): Extract<DeleteTarget, { type: "file" }> {
  const dependents = resourceName
    ? getDependentFiles(path, resourceName, resources)
    : [];
  const kind = NAMED_KINDS[resourceName?.kind as ResourceKind];
  if (!kind || !resourceName?.name) return { type: "file", path, dependents };
  return {
    type: "file",
    path,
    resource: { kind, name: resourceName.name },
    dependents,
  };
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

/**
 * Returns the files of the resources that depend on the given one, directly or through other resources, sorted by path.
 * Canvas components are followed but not listed, since a canvas refers to its components.
 * Files are left out when they are the one being deleted, such as an explore defined in its metrics view's file.
 */
function getDependentFiles(
  path: string,
  resourceName: V1ResourceName,
  resources: V1Resource[],
): string[] {
  // Refs keep the name as written, while the runtime matches names case-insensitively.
  const key = (name: V1ResourceName | undefined) =>
    `${name?.kind}/${name?.name?.toLowerCase()}`;

  const dependentsByRef = new Map<string, V1Resource[]>();
  for (const resource of resources) {
    for (const ref of resource.meta?.refs ?? []) {
      const dependents = dependentsByRef.get(key(ref)) ?? [];
      dependents.push(resource);
      dependentsByRef.set(key(ref), dependents);
    }
  }

  const visited = new Set([key(resourceName)]);
  const queue = [key(resourceName)];
  const files = new Set<string>();
  while (queue.length) {
    for (const dependent of dependentsByRef.get(queue.shift()!) ?? []) {
      const dependentKey = key(dependent.meta?.name);
      if (visited.has(dependentKey)) continue;
      visited.add(dependentKey);
      queue.push(dependentKey);
      if (dependent.meta?.name?.kind === ResourceKind.Component) continue;
      for (const file of dependent.meta?.filePaths ?? []) {
        if (file !== path) files.add(file);
      }
    }
  }
  return [...files].sort();
}
