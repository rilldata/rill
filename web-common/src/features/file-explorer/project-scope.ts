import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { createLocalServiceGetMetadata } from "@rilldata/web-common/runtime-client/local-service";
import { derived, get, readable, type Readable } from "svelte/store";

/**
 * Returns a store with the id that scopes the file explorer's persisted expand/collapse state to the current project.
 * It holds undefined until the id is known.
 *
 * In Rill Cloud the instanceId is unique per project.
 * In Rill Developer it is always "default", so the project's on-disk path is used instead.
 * The path comes from GetMetadata, a Rill Developer endpoint that Rill Cloud answers with 501 Not Implemented,
 * so it is only queried in Rill Developer.
 *
 * Call it during component initialization, since it creates a query.
 */
export function useProjectScopeId(
  instanceId: string,
): Readable<string | undefined> {
  // adminServer defaults to false and Rill Cloud's root layout sets it before any child mounts; it never changes afterwards.
  if (get(featureFlags.adminServer)) return readable(instanceId);

  const metadataQuery = createLocalServiceGetMetadata({
    query: { retry: false },
  });
  return derived(metadataQuery, ($metadataQuery) =>
    $metadataQuery.isLoading
      ? undefined
      : ($metadataQuery.data?.projectPath ?? instanceId),
  );
}
