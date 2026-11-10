import {
  adminServiceGetDeploymentCredentials,
  adminServiceGetProject,
  getAdminServiceGetDeploymentCredentialsQueryKey,
  getAdminServiceGetProjectQueryKey,
  V1DeploymentStatus,
  type RpcStatus,
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
import type { CreateQueryOptions, QueryKey } from "@tanstack/svelte-query";

const PollTimeWhenProjectDeploymentPending = 1000;
const PollTimeWhenProjectDeploymentError = 5000;
const PollTimeWhenProjectDeploymentOk = RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires
// Retries of a JWT refetch that failed with a network or server error; delays come from the query client's retryDelay.
const FreshJwtMaxRetries = 3;
// A JWT refetch attempt still pending after this long is aborted and retried; a stalled request can take minutes to fail.
const FreshJwtAttemptTimeout = 15_000;

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
      return () =>
        fetchFreshJwt(
          getAdminServiceGetDeploymentCredentialsQueryKey(
            organization,
            project,
            body,
          ),
          (signal) =>
            adminServiceGetDeploymentCredentials(
              organization,
              project,
              body,
              signal,
            ),
          (data) => data.accessToken,
        );
    }
    case "magic": {
      if (!token) return undefined;
      return () =>
        fetchFreshJwt(
          getAdminServiceGetProjectWithBearerTokenQueryKey(
            organization,
            project,
            token,
          ),
          (signal) =>
            adminServiceGetProjectWithBearerToken(
              organization,
              project,
              token,
              undefined,
              signal,
            ),
          (data) => data.jwt,
        );
    }
    case "user": {
      const params = branch ? { branch } : undefined;
      return () =>
        fetchFreshJwt(
          getAdminServiceGetProjectQueryKey(organization, project, params),
          (signal) =>
            adminServiceGetProject(organization, project, params, signal),
          (data) => data.jwt,
        );
    }
    case "embed":
      // Embed tokens come from the host page and last 24 hours; there is nothing to refetch.
      return undefined;
  }
}

/**
 * Fetches the query that issues a JWT, through the query client: concurrent callers and a refetch already in flight
 * share one request, and network or server errors are retried with backoff. Refusals (4xx) fail at once.
 * Each attempt is cut off after FreshJwtAttemptTimeout, so a stalled request is retried instead of being joined by every
 * later caller. No caller cancels the shared request; one that stops waiting simply stops listening.
 */
export function fetchFreshJwt<TData>(
  queryKey: QueryKey,
  fetchData: (signal: AbortSignal) => Promise<TData>,
  selectJwt: (data: TData) => string | undefined,
): Promise<string | undefined> {
  return queryClient
    .fetchQuery<TData, RpcStatus>({
      queryKey,
      queryFn: ({ signal }) =>
        fetchData(
          AbortSignal.any([
            signal,
            AbortSignal.timeout(FreshJwtAttemptTimeout),
          ]),
        ),
      // Always refetch unless a request is already in flight: the cached token may be the one the runtime just rejected
      staleTime: 0,
      retry: (failureCount, error) => {
        const status = extractErrorStatusCode(error);
        return (
          failureCount < FreshJwtMaxRetries &&
          (status === undefined || status >= 500)
        );
      },
    })
    .then(selectJwt);
}
