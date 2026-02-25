import type { V1Resource } from "../../runtime-client";
import { createRuntimeServiceListResources } from "../../runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "../../runtime-client/v2";
import { resourceNameToId } from "./resource-utils";

/**
 * Query factory that fetches all resources (only when there's an error)
 * and resolves the root cause error message via `select`.
 *
 * Only use in Rill Developer. Root cause errors expose dependency graph
 * details that are useful for editors and admins but not appropriate for
 * Cloud viewers, who have no access to the project's resource graph.
 */
export function createRootCauseErrorQuery(
  client: RuntimeClient,
  resource: V1Resource | undefined,
  errorMessage: string | undefined,
) {
  return createRuntimeServiceListResources<string | undefined>(
    client,
    {},
    {
      query: {
        enabled: !!errorMessage && !!resource,
        select: (data) =>
          resolveRootCauseErrorMessage(
            resource!,
            data.resources ?? [],
            errorMessage!,
          ),
      },
    },
  );
}

/**
 * Resolves the root cause error message for a resource. If the resource's
 * error is caused by a dependency, traverses refs to find the original
 * error and formats it as "Error in dependency resourceName: error". Returns the original
 * error message if no deeper cause is found.
 */
export function resolveRootCauseErrorMessage(
  resource: V1Resource,
  allResources: V1Resource[],
  errorMessage: string,
): string {
  const rootCause = findRootCause(resource, allResources);
  if (rootCause?.meta?.reconcileError) {
    return `Error in dependency ${rootCause.meta.name?.name}: ${rootCause.meta.reconcileError}`;
  }
  return errorMessage;
}

/**
 * Given a resource with a dependency error, traverse meta.refs to find the
 * root cause resource — the one whose error is NOT caused by another dependency.
 */
function findRootCause(
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
