import type { V1Resource, V1ResourceName } from "../../runtime-client";
import { resourceNameToId } from "./resource-utils";

/**
 * Given a resource with a dependency error, traverse meta.refs
 * to find the root cause resource (the one whose error is NOT
 * caused by another dependency).
 */
export function findRootCause(
  resource: V1Resource,
  allResources: V1Resource[],
): V1Resource | undefined {
  const refs = resource.meta?.refs;
  if (!refs?.length) return undefined;

  const erroredRef = findErroredRef(refs, allResources);
  if (!erroredRef) return undefined;

  // If that ref also has errored refs, keep traversing
  const deeper = findRootCause(erroredRef, allResources);
  return deeper ?? erroredRef;
}

function findErroredRef(
  refs: V1ResourceName[],
  allResources: V1Resource[],
): V1Resource | undefined {
  for (const ref of refs) {
    const refId = resourceNameToId(ref);
    const refResource = allResources.find(
      (r) => resourceNameToId(r.meta?.name) === refId,
    );
    if (refResource?.meta?.reconcileError) {
      return refResource;
    }
  }
  return undefined;
}
