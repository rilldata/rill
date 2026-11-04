import {
  adminServiceGetDeploymentCredentials,
  adminServiceGetProject,
  getAdminServiceGetDeploymentCredentialsQueryKey,
  getAdminServiceGetProjectQueryKey,
  V1DeploymentStatus,
  type RpcStatus,
  type V1GetDeploymentCredentialsResponse,
  type V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import {
  adminServiceGetProjectWithBearerToken,
  getAdminServiceGetProjectWithBearerTokenQueryKey,
} from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
import { extractErrorStatusCode } from "@rilldata/web-common/lib/errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
import type {
  AuthContext,
  JwtRefresher,
} from "@rilldata/web-common/runtime-client/v2";
import type {
  CreateQueryOptions,
  FetchQueryOptions,
} from "@tanstack/svelte-query";

const PollTimeWhenProjectDeploymentPending = 1000;
const PollTimeWhenProjectDeploymentError = 5000;
const PollTimeWhenProjectDeploymentOk = RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires
// A refetch that completed this recently already carries a fresh JWT; hand it over instead of fetching again.
const FreshJwtStaleTime = 5000;

export const baseGetProjectQueryOptions: Partial<
  CreateQueryOptions<V1GetProjectResponse, RpcStatus>
> = {
  gcTime: Math.min(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL, 1000 * 60 * 5), // Make sure we don't keep a stale JWT in the cache
  refetchInterval: (query) => {
    // The session or the project access is gone: polling would repeat the same 401/403 until the tab is closed.
    const errorStatus = extractErrorStatusCode(query.state.error);
    if (errorStatus === 401 || errorStatus === 403) return false;

    const status = query.state.data?.deployment?.status;
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING:
        return PollTimeWhenProjectDeploymentPending;
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
        return PollTimeWhenProjectDeploymentError;
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
        return PollTimeWhenProjectDeploymentOk;
      default:
        return false;
    }
  },
  refetchIntervalInBackground: true, // Keep polling while the tab is hidden (e.g. deploy loader)
  refetchOnMount: true,
  refetchOnReconnect: true,
  refetchOnWindowFocus: true,
  // Align staleness with the RUNNING poll cadence so near-simultaneous observer
  // mounts (e.g. project + edit layouts) dedupe instead of each firing its own
  // refetch. Focus/reconnect refetch still kicks in once data exceeds this window.
  staleTime: PollTimeWhenProjectDeploymentOk,
};

/**
 * Creates the callback `RuntimeProvider` uses to fetch a fresh JWT when a request finds the current one
 * expired, which happens when a laptop sleep or a network drop skipped the scheduled refetch above.
 * It re-runs the admin query that issued the JWT for the active auth mode, keyed exactly like the
 * layout's own query so a refetch here and a scheduled one dedupe.
 */
export function createRuntimeJwtRefresher(
  authContext: AuthContext,
  opts: {
    organization: string;
    project: string;
    branch?: string;
    /** Magic auth token of the public URL being viewed. */
    token?: string;
    /** User impersonated with View As. */
    mockedUserId?: string;
  },
): JwtRefresher | undefined {
  const { organization, project, branch, token, mockedUserId } = opts;
  switch (authContext) {
    case "mock": {
      if (!mockedUserId) return undefined;
      const body = { userId: mockedUserId, ...(branch ? { branch } : {}) };
      return (signal) =>
        fetchFreshJwt<V1GetDeploymentCredentialsResponse>(
          {
            queryKey: getAdminServiceGetDeploymentCredentialsQueryKey(
              organization,
              project,
              body,
            ),
            queryFn: ({ signal }) =>
              adminServiceGetDeploymentCredentials(
                organization,
                project,
                body,
                signal,
              ),
          },
          (data) => data.accessToken,
          signal,
        );
    }
    case "magic": {
      if (!token) return undefined;
      return (signal) =>
        fetchFreshJwt<V1GetProjectResponse>(
          {
            queryKey: getAdminServiceGetProjectWithBearerTokenQueryKey(
              organization,
              project,
              token,
            ),
            queryFn: ({ signal }) =>
              adminServiceGetProjectWithBearerToken(
                organization,
                project,
                token,
                undefined,
                signal,
              ),
          },
          (data) => data.jwt,
          signal,
        );
    }
    case "user": {
      const params = branch ? { branch } : undefined;
      return (signal) =>
        fetchFreshJwt<V1GetProjectResponse>(
          {
            queryKey: getAdminServiceGetProjectQueryKey(
              organization,
              project,
              params,
            ),
            queryFn: ({ signal }) =>
              adminServiceGetProject(organization, project, params, signal),
          },
          (data) => data.jwt,
          signal,
        );
    }
    case "embed":
      // Embed tokens come from the host page and last 24 hours; there is nothing to refetch.
      return undefined;
  }
}

function fetchFreshJwt<TData>(
  options: FetchQueryOptions<TData, RpcStatus>,
  selectJwt: (data: TData) => string | undefined,
  signal: AbortSignal,
): Promise<string | undefined> {
  // The client abandons a call that stalls; cancel the request so the retry is not deduped into it
  const cancel = () =>
    void queryClient.cancelQueries({ queryKey: options.queryKey });
  signal.addEventListener("abort", cancel, { once: true });
  return queryClient
    .fetchQuery({ ...options, staleTime: FreshJwtStaleTime })
    .then(selectJwt)
    .finally(() => signal.removeEventListener("abort", cancel));
}
