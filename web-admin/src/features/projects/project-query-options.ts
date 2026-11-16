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
import type {
  CreateQueryOptions,
  QueryFunctionContext,
  QueryKey,
} from "@tanstack/svelte-query";

const PollTimeWhenProjectDeploymentPending = 1000;
const PollTimeWhenProjectDeploymentError = 5000;
const PollTimeWhenProjectDeploymentOk = RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires
// Retries of a request for a JWT that failed with a network or server error; delays come from the query client's retryDelay.
const JwtRequestMaxRetries = 3;
// A request for a JWT still pending after this long is aborted and retried; a stalled request can take minutes to fail.
// Four attempts and the query client's retry delays (1, 2 and 4 seconds) take at most 55 seconds, so a refresh that
// keeps stalling fails before the requests waiting for it give up (JWT_REFRESH_TIMEOUT).
const JwtRequestAttemptTimeout = 12_000;

const baseGetProjectQueryOptions: Partial<
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
 * Options for the layouts' GetProject queries, which poll for the deployment and its runtime JWT.
 * `fetchProject` is the query's request, passed again so it gets the deadline and retries of a JWT refresh:
 * a refresh joins a refetch already in flight, so a poll stalled across a laptop sleep would hold it up too.
 */
export function getProjectQueryOptions(
  fetchProject: (signal: AbortSignal) => Promise<V1GetProjectResponse>,
): Partial<CreateQueryOptions<V1GetProjectResponse, RpcStatus>> {
  return { ...baseGetProjectQueryOptions, ...jwtRequestOptions(fetchProject) };
}

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
 * share one request, which gets the deadline and retries of jwtRequestOptions.
 * No caller cancels the shared request; one that stops waiting simply stops listening.
 */
export function fetchFreshJwt<TData>(
  queryKey: QueryKey,
  fetchData: (signal: AbortSignal) => Promise<TData>,
  selectJwt: (data: TData) => string | undefined,
): Promise<string | undefined> {
  return queryClient
    .fetchQuery<TData, RpcStatus>({
      queryKey,
      ...jwtRequestOptions(fetchData),
      // Always refetch unless a request is already in flight: the cached token may be the one the runtime just rejected
      staleTime: 0,
    })
    .then(selectJwt);
}

/**
 * Request and retry options shared by the queries that issue a runtime JWT, whether a layout polls them or a refresh
 * fetches them on demand. Each attempt is cut off after JwtRequestAttemptTimeout and retried: a request stalled by a
 * laptop sleep or a network change can take minutes to fail, and every refetch in the meantime would join it.
 * Network and server errors are retried with backoff; refusals (4xx) fail at once.
 */
function jwtRequestOptions<TData>(
  fetchData: (signal: AbortSignal) => Promise<TData>,
) {
  return {
    queryFn: async ({ signal }: QueryFunctionContext) => {
      // Aborts when the query is cancelled or the attempt runs out of time. AbortSignal.any would need Safari 17.4.
      const attempt = new AbortController();
      const cancel = () => attempt.abort(signal.reason);
      signal.addEventListener("abort", cancel);
      const deadline = setTimeout(
        () =>
          attempt.abort(
            new DOMException("The JWT request timed out", "TimeoutError"),
          ),
        JwtRequestAttemptTimeout,
      );
      try {
        return await fetchData(attempt.signal);
      } finally {
        clearTimeout(deadline);
        signal.removeEventListener("abort", cancel);
      }
    },
    retry: (failureCount: number, error: RpcStatus) => {
      const status = extractErrorStatusCode(error);
      return (
        failureCount < JwtRequestMaxRetries &&
        (status === undefined || status >= 500)
      );
    },
  };
}
