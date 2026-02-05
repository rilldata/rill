import type { PartialMessage } from "@bufbuild/protobuf";
import { type ConnectError, createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { LocalService } from "@rilldata/web-common/proto/gen/rill/local/v1/api_connect";
import {
  DeployProjectRequest,
  GetCurrentProjectRequest,
  GetCurrentUserRequest,
  GetMetadataRequest,
  ListOrganizationsAndBillingMetadataRequest,
  GetVersionRequest,
  PushToGithubRequest,
  RedeployProjectRequest,
  CreateOrganizationRequest,
  ListMatchingProjectsRequest,
  ListProjectsForOrgRequest,
  GetProjectRequest,
  GitStatusRequest,
  GitPullRequest,
  GitPushRequest,
  GithubRepoStatusRequest,
} from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  createMutation,
  createQuery,
  type CreateMutationOptions,
  type CreateQueryOptions,
  type QueryFunction,
  type DataTag,
  type QueryKey,
} from "@tanstack/svelte-query";
import { get } from "svelte/store";

/**
 * Handwritten wrapper on LocalService.
 * TODO: find a way to autogenerate
 */

// cache of clients per host
const clients = new Map<
  string,
  ReturnType<typeof createPromiseClient<typeof LocalService>>
>();
function getClient() {
  const host = get(runtime).host;
  if (clients.has(host)) return clients.get(host)!;

  const transport = createConnectTransport({
    baseUrl: host,
  });
  const client = createPromiseClient(LocalService, transport);
  clients.set(host, client);
  return client;
}

export function localServiceGetMetadata() {
  return getClient().getMetadata(new GetMetadataRequest());
}
export const getLocalServiceGetMetadataQueryKey = () => [
  `/v1/local/get-metadata`,
];
export function createLocalServiceGetMetadata<
  TData = Awaited<ReturnType<typeof localServiceGetMetadata>>,
  TError = ConnectError,
>(options?: {
  query?: Partial<
    CreateQueryOptions<
      Awaited<ReturnType<typeof localServiceGetMetadata>>,
      TError,
      TData
    >
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey: queryOptions?.queryKey ?? getLocalServiceGetMetadataQueryKey(),
    queryFn: queryOptions?.queryFn ?? localServiceGetMetadata,
  });
}

export function localServiceGetVersion() {
  return getClient().getVersion(new GetVersionRequest());
}
export const getLocalServiceGetVersionQueryKey = () => [
  `/v1/local/get-version`,
];
export function createLocalServiceGetVersion<
  TData = Awaited<ReturnType<typeof localServiceGetVersion>>,
  TError = ConnectError,
>(options?: {
  query?: CreateQueryOptions<
    Awaited<ReturnType<typeof localServiceGetVersion>>,
    TError,
    TData
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey: queryOptions?.queryKey ?? getLocalServiceGetVersionQueryKey(),
    queryFn: queryOptions?.queryFn ?? localServiceGetVersion,
  });
}

export function localServicePushToGithub(
  args: PartialMessage<PushToGithubRequest>,
) {
  return getClient().pushToGithub(new PushToGithubRequest(args));
}
export function createLocalServicePushToGithub<
  TError = ConnectError,
  TContext = unknown,
>(options?: {
  mutation?: Partial<
    CreateMutationOptions<
      Awaited<ReturnType<typeof localServicePushToGithub>>,
      TError,
      PartialMessage<PushToGithubRequest>,
      TContext
    >
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServicePushToGithub>>,
    unknown,
    PartialMessage<PushToGithubRequest>,
    unknown
  >({ mutationFn: localServicePushToGithub, ...mutationOptions });
}

export function localServiceDeploy(args: PartialMessage<DeployProjectRequest>) {
  return getClient().deployProject(new DeployProjectRequest(args));
}
export function createLocalServiceDeploy<
  TError = ConnectError,
  TContext = unknown,
>(options?: {
  mutation?: Partial<
    CreateMutationOptions<
      Awaited<ReturnType<typeof localServiceDeploy>>,
      TError,
      PartialMessage<DeployProjectRequest>,
      TContext
    >
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServiceDeploy>>,
    unknown,
    PartialMessage<DeployProjectRequest>,
    unknown
  >({ mutationFn: localServiceDeploy, ...mutationOptions });
}

export function localServiceRedeploy(
  args: PartialMessage<RedeployProjectRequest>,
) {
  return getClient().redeployProject(new RedeployProjectRequest(args));
}
export function createLocalServiceRedeploy<
  TError = ConnectError,
  TContext = unknown,
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof localServiceRedeploy>>,
    TError,
    PartialMessage<RedeployProjectRequest>,
    TContext
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServiceRedeploy>>,
    unknown,
    PartialMessage<RedeployProjectRequest>,
    unknown
  >({ mutationFn: localServiceRedeploy, ...mutationOptions });
}

export function localServiceGetCurrentUser() {
  return getClient().getCurrentUser(new GetCurrentUserRequest());
}
export const getLocalServiceGetCurrentUserQueryKey = () => [
  `/v1/local/get-user`,
];
export function createLocalServiceGetCurrentUser<
  TData = Awaited<ReturnType<typeof localServiceGetCurrentUser>>,
  TError = ConnectError,
>(options?: {
  query?: Partial<
    CreateQueryOptions<
      Awaited<ReturnType<typeof localServiceGetCurrentUser>>,
      TError,
      TData
    >
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey: queryOptions?.queryKey ?? getLocalServiceGetCurrentUserQueryKey(),
    queryFn: queryOptions?.queryFn ?? localServiceGetCurrentUser,
  });
}

export function localServiceGetCurrentProject() {
  return getClient().getCurrentProject(new GetCurrentProjectRequest());
}
export const getLocalServiceGetCurrentProjectQueryKey = () => [
  `/v1/local/get-project`,
];
export function createLocalServiceGetCurrentProject<
  TData = Awaited<ReturnType<typeof localServiceGetCurrentProject>>,
  TError = ConnectError,
>(options?: {
  query?: Partial<
    CreateQueryOptions<
      Awaited<ReturnType<typeof localServiceGetCurrentProject>>,
      TError,
      TData
    >
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey:
      queryOptions?.queryKey ?? getLocalServiceGetCurrentProjectQueryKey(),
    queryFn: queryOptions?.queryFn ?? localServiceGetCurrentProject,
  });
}

export function localServiceListOrganizationsAndBillingMetadataRequest() {
  return getClient().listOrganizationsAndBillingMetadata(
    new ListOrganizationsAndBillingMetadataRequest(),
  );
}
export const getLocalServiceListOrganizationsAndBillingMetadataRequestQueryKey =
  () => [`/v1/local/list-organizations-billing-metadata`];
export function createLocalServiceListOrganizationsAndBillingMetadataRequest<
  TData = Awaited<
    ReturnType<typeof localServiceListOrganizationsAndBillingMetadataRequest>
  >,
  TError = ConnectError,
>(options?: {
  query?: Partial<
    CreateQueryOptions<
      Awaited<
        ReturnType<
          typeof localServiceListOrganizationsAndBillingMetadataRequest
        >
      >,
      TError,
      TData
    >
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey:
      queryOptions?.queryKey ??
      getLocalServiceListOrganizationsAndBillingMetadataRequestQueryKey(),
    queryFn:
      queryOptions?.queryFn ??
      localServiceListOrganizationsAndBillingMetadataRequest,
  });
}

export function localServiceCreateOrganization(
  args: PartialMessage<CreateOrganizationRequest>,
) {
  return getClient().createOrganization(new CreateOrganizationRequest(args));
}
export function createLocalServiceCreateOrganization<
  TError = ConnectError,
  TContext = unknown,
>(options?: {
  mutation?: Partial<
    CreateMutationOptions<
      Awaited<ReturnType<typeof localServiceCreateOrganization>>,
      TError,
      PartialMessage<CreateOrganizationRequest>,
      TContext
    >
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServiceCreateOrganization>>,
    unknown,
    PartialMessage<CreateOrganizationRequest>,
    unknown
  >({ mutationFn: localServiceCreateOrganization, ...mutationOptions });
}

export function localServiceListMatchingProjectsRequest() {
  return getClient().listMatchingProjects(new ListMatchingProjectsRequest());
}
export const getLocalServiceListMatchingProjectsRequestQueryKey = () => [
  `/v1/local/list-matching-projects`,
];
export function createLocalServiceListMatchingProjectsRequest<
  TData = Awaited<ReturnType<typeof localServiceListMatchingProjectsRequest>>,
  TError = ConnectError,
>(options?: {
  query?: Partial<
    CreateQueryOptions<
      Awaited<ReturnType<typeof localServiceListMatchingProjectsRequest>>,
      TError,
      TData
    >
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey:
      queryOptions?.queryKey ??
      getLocalServiceListMatchingProjectsRequestQueryKey(),
    queryFn: queryOptions?.queryFn ?? localServiceListMatchingProjectsRequest,
  });
}

export function localServiceListProjectsForOrgRequest(org: string) {
  return getClient().listProjectsForOrg(
    new ListProjectsForOrgRequest({
      org,
    }),
  );
}
export const getLocalServiceListProjectsForOrgRequestQueryKey = (
  org: string,
) => [`/v1/local/list-projects-for-org`, org];
export function createLocalServiceListProjectsForOrgRequest<
  TData = Awaited<ReturnType<typeof localServiceListProjectsForOrgRequest>>,
  TError = ConnectError,
>(
  org: string,
  options?: {
    query?: Partial<
      CreateQueryOptions<
        Awaited<ReturnType<typeof localServiceListProjectsForOrgRequest>>,
        TError,
        TData
      >
    >;
  },
) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey:
      queryOptions?.queryKey ??
      getLocalServiceListProjectsForOrgRequestQueryKey(org),
    queryFn:
      queryOptions?.queryFn ??
      (() => localServiceListProjectsForOrgRequest(org)),
  });
}

export function localServiceGitStatus() {
  return getClient().gitStatus(new GitStatusRequest({}));
}
export const getLocalServiceGitStatusQueryKey = () => [`/v1/local/git-status`];
export function createLocalServiceGitStatus<
  TData = Awaited<ReturnType<typeof localServiceGitStatus>>,
  TError = ConnectError,
>(options?: {
  query?: Partial<
    CreateQueryOptions<
      Awaited<ReturnType<typeof localServiceGitStatus>>,
      TError,
      TData
    >
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey: queryOptions?.queryKey ?? getLocalServiceGitStatusQueryKey(),
    queryFn: queryOptions?.queryFn ?? (() => localServiceGitStatus()),
  });
}

export function localServiceGithubRepoStatus(remote: string) {
  return getClient().githubRepoStatus(
    new GithubRepoStatusRequest({
      remote,
    }),
  );
}
export const getLocalServiceGithubRepoStatusQueryKey = (remote: string) => [
  `/v1/local/git-repo-status`,
  remote,
];
export const getLocalServiceGithubRepoStatusQueryOptions = <
  TData = Awaited<ReturnType<typeof localServiceGithubRepoStatus>>,
  TError = ConnectError,
>(
  remote: string,
  options?: {
    query?: Partial<
      CreateQueryOptions<
        Awaited<ReturnType<typeof localServiceGithubRepoStatus>>,
        TError,
        TData
      >
    >;
  },
) => {
  const { query: queryOptions } = options ?? {};

  const queryKey =
    queryOptions?.queryKey ?? getLocalServiceGithubRepoStatusQueryKey(remote);

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof localServiceGithubRepoStatus>>
  > = () => localServiceGithubRepoStatus(remote);

  return { queryKey, queryFn, ...queryOptions } as CreateQueryOptions<
    Awaited<ReturnType<typeof localServiceGithubRepoStatus>>,
    TError,
    TData
  > & { queryKey: DataTag<QueryKey, TData, TError> };
};
export function createLocalServiceGithubRepoStatus<
  TData = Awaited<ReturnType<typeof localServiceGithubRepoStatus>>,
  TError = ConnectError,
>(
  remote: string,
  options?: {
    query?: Partial<
      CreateQueryOptions<
        Awaited<ReturnType<typeof localServiceGithubRepoStatus>>,
        TError,
        TData
      >
    >;
  },
) {
  const queryOptions = getLocalServiceGithubRepoStatusQueryOptions(
    remote,
    options,
  );
  return createQuery(queryOptions);
}

export function localServiceGitPull(args: PartialMessage<GitPullRequest>) {
  return getClient().gitPull(new GitPullRequest(args));
}
export function createLocalServiceGitPull<
  TError = ConnectError,
  TContext = unknown,
>(options?: {
  mutation?: Partial<
    CreateMutationOptions<
      Awaited<ReturnType<typeof localServiceGitPull>>,
      TError,
      PartialMessage<GitPullRequest>,
      TContext
    >
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServiceGitPull>>,
    TError,
    PartialMessage<GitPullRequest>,
    unknown
  >({ mutationFn: localServiceGitPull, ...mutationOptions });
}

export function localServiceGitPush(args: PartialMessage<GitPushRequest>) {
  return getClient().gitPush(new GitPushRequest(args));
}
export function createLocalServiceGitPush<
  TError = ConnectError,
  TContext = unknown,
>(options?: {
  mutation?: Partial<
    CreateMutationOptions<
      Awaited<ReturnType<typeof localServiceGitPush>>,
      TError,
      PartialMessage<GitPushRequest>,
      TContext
    >
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServiceGitPush>>,
    TError,
    PartialMessage<GitPushRequest>,
    unknown
  >({ mutationFn: localServiceGitPush, ...mutationOptions });
}

export function localServiceGetProjectRequest(
  organizationName: string,
  name: string,
) {
  return getClient().getProject(
    new GetProjectRequest({
      organizationName,
      name,
    }),
  );
}
export const getLocalServiceGetProjectRequestQueryKey = (
  organizationName: string,
  name: string,
) => [`/v1/local/get-project`, organizationName, name];
export function createLocalServiceGetProjectRequest<
  TData = Awaited<ReturnType<typeof localServiceGetProjectRequest>>,
  TError = ConnectError,
>(
  organizationName: string,
  name: string,
  options?: {
    query?: Partial<
      CreateQueryOptions<
        Awaited<ReturnType<typeof localServiceGetProjectRequest>>,
        TError,
        TData
      >
    >;
  },
) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey:
      queryOptions?.queryKey ??
      getLocalServiceGetProjectRequestQueryKey(organizationName, name),
    queryFn:
      queryOptions?.queryFn ??
      (() => localServiceGetProjectRequest(organizationName, name)),
  });
}
