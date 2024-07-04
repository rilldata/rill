import { ConnectError } from "@connectrpc/connect";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { DeployValidationResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  createLocalServiceDeploy,
  createLocalServiceDeployValidation,
  createLocalServiceRedeploy,
} from "@rilldata/web-common/runtime-client/local-service";
import { derived, get, writable } from "svelte/store";

export class ProjectDeployer {
  public readonly validation: ReturnType<
    typeof createLocalServiceDeployValidation
  > = createLocalServiceDeployValidation({
    query: {
      refetchOnWindowFocus: true,
    },
  });
  public readonly deploying = writable(false);

  private readonly deployMutation = createLocalServiceDeploy();
  private readonly redeployMutation = createLocalServiceRedeploy();

  public getStatus() {
    return derived(
      [this.validation, this.deployMutation, this.redeployMutation],
      ([validation, deployMutation, redeployMutation]) => {
        if (
          validation.isFetching ||
          deployMutation.isLoading ||
          redeployMutation.isLoading
        ) {
          return {
            isLoading: true,
            error: undefined,
          };
        }

        return {
          isLoading: false,
          error:
            (validation.error as ConnectError)?.message ??
            (deployMutation.error as ConnectError)?.message ??
            (redeployMutation.error as ConnectError)?.message,
        };
      },
    );
  }

  public get isDeployed() {
    const validation = get(this.validation).data as DeployValidationResponse;
    return !!validation?.deployedProjectId;
  }

  public async validate() {
    this.deploying.set(false);
    let validation = get(this.validation).data as DeployValidationResponse;
    if (validation?.deployedProjectId) {
      return true;
    }

    await waitUntil(() => !get(this.validation).isFetching);
    validation = get(this.validation).data as DeployValidationResponse;

    if (!validation.isAuthenticated) {
      window.open(`${validation.loginUrl}`, "__target");
      this.deploying.set(true);
      return false;
    }

    if (validation.isGithubRepo && !validation.isGithubConnected) {
      // if the project is a github repo and not connected to github then redirect to grant access
      window.open(`${validation.githubGrantAccessUrl}`, "__target");
      this.deploying.set(true);
      return false;
    }

    return true;
  }

  public async deploy() {
    // safeguard around deploy
    if (!(await this.validate())) return;

    const validation = get(this.validation).data as DeployValidationResponse;
    if (validation.deployedProjectId) {
      const resp = await get(this.redeployMutation).mutateAsync({
        projectId: validation.deployedProjectId,
        reupload: !validation.isGithubRepo,
      });
      window.open(resp.frontendUrl, "__target");
    } else {
      const resp = await get(this.deployMutation).mutateAsync({
        projectName: validation.localProjectName,
        org: validation.rillUserOrgs[0],
        upload: !validation.isGithubRepo,
      });
      window.open(resp.frontendUrl, "__target");
    }
  }
}
