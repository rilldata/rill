import type { QueryFunction, QueryKey } from "@tanstack/svelte-query";
import {
  adminServiceGetProject,
  createAdminServiceGetProject,
  createAdminServiceListProjectMemberUsers,
  getAdminServiceGetProjectQueryKey,
  getAdminServiceGetProjectQueryOptions,
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
  ResourceKind,
  SingletonProjectParserName,
  useResourceV2,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
