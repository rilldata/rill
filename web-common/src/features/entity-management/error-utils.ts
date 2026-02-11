import type { V1Resource } from "../../runtime-client";
import { resourceNameToId } from "./resource-utils";

/**
 * Resolves the root cause error message for a resource. If the resource's
 * error is caused by a dependency, traverses refs to find the original
 * error and formats it as "resourceName: error". Falls back to the
 * provided error message if no deeper cause is found.
 */
export function resolveRootCauseErrorMessage(
  resource: V1Resource,
  allResources: V1Resource[],
  fallback: string,
): string {
  const rootCause = findRootCause(resource, allResources);
  if (rootCause?.meta?.reconcileError) {
    return `${rootCause.meta.name?.name}: ${rootCause.meta.reconcileError}`;
  }
  return fallback;
}

/**
 * Given a resource with a dependency error, traverse meta.refs to find the
 * root cause resource — the one whose error is NOT caused by another dependency.
 */
export function findRootCause(
  resource: V1Resource,
  allResources: V1Resource[],
): V1Resource | undefined {
  const refs = resource.meta?.refs;
  if (!refs?.length) return undefined;

  // Find the first ref that has an error
  for (const ref of refs) {
    const refId = resourceNameToId(ref);
    const refResource = allResources.find(
      (r) => resourceNameToId(r.meta?.name) === refId,
    );
    if (refResource?.meta?.reconcileError) {
      // If that ref also has errored refs, keep traversing
      return findRootCause(refResource, allResources) ?? refResource;
    }
  }

  return undefined;
}
