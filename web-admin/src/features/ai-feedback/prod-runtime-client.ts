import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
import { baseGetProjectQueryOptions } from "@rilldata/web-admin/features/projects/project-query-options";
import {
  getRuntimeClient,
  type RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";

export interface ProdRuntimeClientState {
  // Cache-managed client for the production deployment's runtime, or null if the
  // project has no running production deployment (or the lookup hasn't resolved).
  client: RuntimeClient | null;
  // False while the deployment lookup is still pending, so "not deployed" states
  // aren't shown prematurely.
  resolved: boolean;
}

// createProdRuntimeClient resolves a RuntimeClient for the project's production
// deployment. GetProject without a branch returns the primary deployment and mints
// a JWT for it (including ManageAIFeedback for project admins); the query keeps
// polling via baseGetProjectQueryOptions, and getRuntimeClient updates the cached
// client's JWT on every refresh. No RuntimeProvider is involved: providers have
// global side effects (feature-flag rebinding, env stores) that must stay bound to
// the ambient edit-session runtime.
export function createProdRuntimeClient(
  organization: string,
  project: string,
): Readable<ProdRuntimeClientState> {
  const query = createAdminServiceGetProject(organization, project, undefined, {
    query: baseGetProjectQueryOptions,
  });
  return derived(query, ($query) => {
    const deployment = $query.data?.deployment;
    const jwt = $query.data?.jwt;
    const client =
      deployment?.runtimeHost && deployment?.runtimeInstanceId && jwt
        ? getRuntimeClient({
            host: deployment.runtimeHost,
            instanceId: deployment.runtimeInstanceId,
            jwt,
          })
        : null;
    return { client, resolved: !$query.isPending };
  });
}
