import type { PartialMessage } from "@bufbuild/protobuf";
import type { ConnectError } from "@connectrpc/connect";
import { createMutation, CreateMutationOptions } from "@rilldata/svelte-query";
import { PushToGithubRequest } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  localServiceDeploy,
  localServiceDeployValidation,
} from "@rilldata/web-common/runtime-client/local-service";

export function createDeployer(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof deploy>>,
    ConnectError,
    PartialMessage<PushToGithubRequest>,
    unknown
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof deploy>>,
    unknown,
    PartialMessage<PushToGithubRequest>,
    unknown
  >(deploy, mutationOptions);
}

async function deploy() {
  const deployValidation = await localServiceDeployValidation();
  if (!deployValidation.isAuthenticated) {
    window.open(`${deployValidation.loginUrl}`, "__target");
    return false;
  }

  if (deployValidation.isGithubRepo && !deployValidation.isGithubConnected) {
    // if the project is a github repo and not connected to github then redirect to grant access
    window.open(`${deployValidation.githubGrantAccessUrl}`, "__target");
    return false;
  }

  const resp = await localServiceDeploy({
    projectName: deployValidation.localProjectName,
    org: deployValidation.rillUserOrgs[0],
    upload: !deployValidation.isGithubRepo,
  });
  if (resp.frontendUrl) {
    window.open(resp.frontendUrl, "__target");
  }
  return true;
}
