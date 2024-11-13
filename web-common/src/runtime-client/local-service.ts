import type { PartialMessage } from "@bufbuild/protobuf";
import { type ConnectError, createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { LocalService } from "@rilldata/web-common/proto/gen/rill/local/v1/api_connect";
import {
  DeployProjectRequest,
  GetCurrentProjectRequest,
  GetCurrentUserRequest,
  GetMetadataRequest,
  GetUserOrgMetadataRequest,
  GetVersionRequest,
  PushToGithubRequest,
  RedeployProjectRequest,
} from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  createMutation,
  createQuery,
  type CreateMutationOptions,
  type CreateQueryOptions,
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
  query?: CreateQueryOptions<
    Awaited<ReturnType<typeof localServiceGetMetadata>>,
    TError,
    TData
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
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof localServicePushToGithub>>,
    TError,
    PartialMessage<PushToGithubRequest>,
    TContext
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServicePushToGithub>>,
    unknown,
    PartialMessage<PushToGithubRequest>,
    unknown
  >(localServicePushToGithub, mutationOptions);
}

export function localServiceDeploy(args: PartialMessage<DeployProjectRequest>) {
  return getClient().deployProject(new DeployProjectRequest(args));
}
export function createLocalServiceDeploy<
  TError = ConnectError,
  TContext = unknown,
>(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof localServiceDeploy>>,
    TError,
    PartialMessage<DeployProjectRequest>,
    TContext
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof localServiceDeploy>>,
    unknown,
    PartialMessage<DeployProjectRequest>,
    unknown
  >(localServiceDeploy, mutationOptions);
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
  >(localServiceRedeploy, mutationOptions);
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
  query?: CreateQueryOptions<
    Awaited<ReturnType<typeof localServiceGetCurrentUser>>,
    TError,
    TData
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
  query?: CreateQueryOptions<
    Awaited<ReturnType<typeof localServiceGetCurrentProject>>,
    TError,
    TData
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

export function localServiceGetUserOrgMetadataRequest() {
  return getClient().getUserOrgMetadata(new GetUserOrgMetadataRequest());
}
export const getLocalServiceGetUserOrgMetadataRequestQueryKey = () => [
  `/v1/local/get-org-metadata`,
];
export function createLocalServiceGetUserOrgMetadataRequest<
  TData = Awaited<ReturnType<typeof localServiceGetUserOrgMetadataRequest>>,
  TError = ConnectError,
>(options?: {
  query?: CreateQueryOptions<
    Awaited<ReturnType<typeof localServiceGetUserOrgMetadataRequest>>,
    TError,
    TData
  >;
}) {
  const { query: queryOptions } = options ?? {};
  return createQuery({
    ...queryOptions,
    queryKey:
      queryOptions?.queryKey ??
      getLocalServiceGetUserOrgMetadataRequestQueryKey(),
    queryFn: queryOptions?.queryFn ?? localServiceGetUserOrgMetadataRequest,
  });
}
