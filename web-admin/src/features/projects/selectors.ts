import {
  createQuery,
  type QueryFunction,
  type QueryKey,
} from "@tanstack/svelte-query";
import {
  adminServiceGetProject,
  createAdminServiceGetProject,
  createAdminServiceListProjectMemberUsers,
  getAdminServiceGetProjectQueryKey,
  getAdminServiceGetProjectQueryOptions,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import {
  adminServiceGetMagicAuthToken,
  getAdminServiceGetMagicAuthTokenQueryKey,
} from "@rilldata/web-admin/features/public-urls/get-magic-auth-token";
import {
  adminServiceGetProjectWithBearerToken,
  getAdminServiceGetProjectWithBearerTokenQueryKey,
} from "@rilldata/web-admin/features/public-urls/get-project-with-bearer-token";
import {
  fetchResource,
  ResourceKind,
  SingletonProjectParserName,
  useResourceV2,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeClient,
  type RuntimeClient,
} from "@rilldata/web-common/runtime-client/v2";
import { derived, type Readable } from "svelte/store";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, undefined, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}

export function useProjectMembersEmails(organization: string, project: string) {
  return createAdminServiceListProjectMemberUsers(
    organization,
    project,
    undefined,
    {
      query: {
        select: (data) => {
          return data.members
            ?.filter((member) => !!member?.userEmail)
            .map((member) => member.userEmail);
        },
      },
    },
  );
}

export function useProjectId(orgName: string, projectName: string) {
  return createAdminServiceGetProject(
    orgName,
    projectName,
    {},
    {
      query: {
        enabled: !!orgName && !!projectName,
        select: (resp) => resp.project?.id,
      },
    },
  );
}

export type OrgAndProjectNameStore = Readable<{
  organization: string;
  project: string;
}>;
export function getProjectIdQueryOptions(
  orgAndProjectNameStore: OrgAndProjectNameStore,
) {
  return derived(orgAndProjectNameStore, ({ organization, project }) =>
    getAdminServiceGetProjectQueryOptions(
      organization,
      project,
      {},
      {
        query: {
          enabled: !!organization && !!project,
          select: (resp) => resp.project?.id,
        },
      },
    ),
  );
}

export function fetchMagicAuthToken(token: string) {
  const queryKey = getAdminServiceGetMagicAuthTokenQueryKey(token);
  const queryFunction: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetMagicAuthToken>>
  > = ({ signal }) => adminServiceGetMagicAuthToken(token, signal);

  return queryClient.fetchQuery({
    queryKey,
    queryFn: queryFunction,
  });
}

export async function fetchProjectDeploymentDetails(
  orgName: string,
  projectName: string,
  token: string | undefined,
) {
  let queryKey: QueryKey;
  let queryFn: QueryFunction<
    Awaited<ReturnType<typeof adminServiceGetProject>>
  >;

  if (token) {
    queryKey = getAdminServiceGetProjectWithBearerTokenQueryKey(
      orgName,
      projectName,
      token,
      {},
    );

    queryFn = ({ signal }) =>
      adminServiceGetProjectWithBearerToken(
        orgName,
        projectName,
        token,
        {},
        signal,
      );
  } else {
    queryKey = getAdminServiceGetProjectQueryKey(orgName, projectName);

    queryFn = ({ signal }) =>
      adminServiceGetProject(orgName, projectName, {}, signal);
  }

  const projResp = await queryClient.fetchQuery({
    queryKey,
    queryFn,
  });

  return {
    projectPermissions: projResp.projectPermissions,
    project: projResp.project,
    runtime: {
      host: projResp.deployment?.runtimeHost ?? "",
      instanceId: projResp.deployment?.runtimeInstanceId ?? "",
      jwt: projResp.jwt
        ? {
            token: projResp.jwt,
            receivedAt: Date.now(),
            authContext: (token ? "magic" : "user") as string,
          }
        : undefined,
    },
  };
}

/**
 * Reactive query for a runtime deployment's project parser commit SHA.
 * Used by Publish/Merge to capture prod's pre-merge state so the
 * deploying page can wait for the parser to advance past it before
 * redirecting. Disabled until the deployment's runtime info is
 * available; on transport errors, logs and resolves to undefined so
 * callers can fall back to the "skip the SHA gate" behavior rather
 * than show a query-error state.
 *
 * Refetches every 5 minutes so the cached value the popover reads at
 * click time stays reasonably current — without this, an editor
 * session left open for an hour while another commit lands on primary
 * could pass a stale SHA and trigger a false-early redirect.
 */
const PARSER_SHA_REFETCH_INTERVAL_MS = 5 * 60 * 1000;
export function useParserCommitSha(
  deployment: V1Deployment | undefined,
  jwt: string | undefined,
) {
  const host = deployment?.runtimeHost;
  const instanceId = deployment?.runtimeInstanceId;
  return createQuery({
    queryKey: ["parserCommitSha", host ?? "", instanceId ?? ""],
    queryFn: async () => {
      if (!host || !instanceId) return undefined;
      try {
        const client = getRuntimeClient({
          host,
          instanceId,
          jwt,
          authContext: "user",
        });
        const resource = await fetchResource(
          queryClient,
          client,
          SingletonProjectParserName,
          ResourceKind.ProjectParser,
        );
        return resource?.projectParser?.state?.currentCommitSha || undefined;
      } catch (err) {
        console.warn("Failed to fetch parser commit SHA:", err);
        return undefined;
      }
    },
    enabled: !!host && !!instanceId,
    refetchInterval: PARSER_SHA_REFETCH_INTERVAL_MS,
  });
}

export function useGithubLastSynced(client: RuntimeClient) {
  return useResourceV2(
    client,
    SingletonProjectParserName,
    ResourceKind.ProjectParser,
    {
      select: (data) =>
        data.resource?.projectParser?.state?.currentCommitOn
          ? new Date(data.resource.projectParser.state.currentCommitOn)
          : null,
    },
  );
}
