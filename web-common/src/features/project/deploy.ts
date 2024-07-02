import type { ConnectError } from "@connectrpc/connect";
import { createMutation, CreateMutationOptions } from "@rilldata/svelte-query";
import { DeployValidationResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  localServiceDeploy,
  localServiceRedeploy,
} from "@rilldata/web-common/runtime-client/local-service";

export function createDeployer(options?: {
  mutation?: CreateMutationOptions<
    Awaited<ReturnType<typeof deploy>>,
    ConnectError,
    DeployValidationResponse,
    unknown
  >;
}) {
  const { mutation: mutationOptions } = options ?? {};
  return createMutation<
    Awaited<ReturnType<typeof deploy>>,
    unknown,
    DeployValidationResponse,
    unknown
  >(deploy, mutationOptions);
}

async function deploy(deployValidation: DeployValidationResponse) {
  if (!deployValidation.isAuthenticated) {
    window.open(`${deployValidation.loginUrl}`, "__target");
    return false;
  }

  if (deployValidation.isGithubRepo && !deployValidation.isGithubConnected) {
    // if the project is a github repo and not connected to github then redirect to grant access
    window.open(`${deployValidation.githubGrantAccessUrl}`, "__target");
    return false;
  }

  if (deployValidation.deployedProjectId) {
    const resp = await localServiceRedeploy({
      projectId: deployValidation.deployedProjectId,
      reupload: !deployValidation.isGithubRepo,
    });
    if (resp.frontendUrl) {
      window.open(resp.frontendUrl, "__target");
    }
  } else {
    const resp = await localServiceDeploy({
      projectName: deployValidation.localProjectName,
      org: deployValidation.rillUserOrgs[0],
      upload: !deployValidation.isGithubRepo,
    });
    if (resp.frontendUrl) {
      window.open(resp.frontendUrl, "__target");
    }
  }
  return true;
}
