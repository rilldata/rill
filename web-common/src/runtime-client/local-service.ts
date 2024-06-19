import { createPromiseClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { LocalService } from "@rilldata/web-common/proto/gen/rill/local/v1/api_connect";
import {
  DeployRequest,
  DeployValidationRequest,
  GetMetadataRequest,
  GetVersionRequest,
  PushToGithubRequest,
} from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { createMutation, createQuery } from "@tanstack/svelte-query";
import { get } from "svelte/store";

/**
 * Handwritten wrapper on LocalService.
 * TODO: Add query and mutation params.
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
  return getClient().deployValidation(new GetMetadataRequest());
}
export const getLocalServiceGetMetadataQueryKey = () => [
  `/v1/local/get-metadata`,
];
export function createLocalServiceGetMetadata() {
  return createQuery({
    queryKey: getLocalServiceGetMetadataQueryKey(),
    queryFn: localServiceGetMetadata,
  });
}

export function localServiceGetVersion() {
  return getClient().deployValidation(new GetVersionRequest());
}
export const getLocalServiceGetVersionQueryKey = () => [
  `/v1/local/get-version`,
];
export function createLocalServiceGetVersion() {
  return createQuery({
    queryKey: getLocalServiceGetVersionQueryKey(),
    queryFn: localServiceGetVersion,
  });
}

export function localServiceDeployValidation() {
  return getClient().deployValidation(new DeployValidationRequest());
}
export const getLocalServiceDeployValidationQueryKey = () => [
  `/v1/local/deploy-validation`,
];
export function createLocalServiceDeployValidation() {
  return createQuery({
    queryKey: getLocalServiceDeployValidationQueryKey(),
    queryFn: localServiceDeployValidation,
  });
}

export function localServicePushToGithub(account: string, repo: string) {
  return getClient().pushToGithub(
    new PushToGithubRequest({
      account,
      repo,
    }),
  );
}
export function createLocalServicePushToGithub() {
  return createMutation<
    Awaited<ReturnType<typeof localServicePushToGithub>>,
    unknown,
    {
      account: string;
      repo: string;
    },
    unknown
  >(({ account, repo }) => localServicePushToGithub(account, repo));
}

export function localServiceDeploy(org: string, projectName: string) {
  return getClient().deploy(
    new DeployRequest({
      org,
      projectName,
    }),
  );
}
export function createLocalServiceDeploy() {
  return createMutation<
    Awaited<ReturnType<typeof localServiceDeploy>>,
    unknown,
    { org: string; projectName: string },
    unknown
  >(({ org, projectName }) => localServiceDeploy(org, projectName));
}
