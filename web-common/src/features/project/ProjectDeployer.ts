import { page } from "$app/stores";
import type { ConnectError } from "@connectrpc/connect";
import { getOrgName } from "@rilldata/web-common/features/project/getOrgName";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import type { DeployValidationResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  createLocalServiceDeploy,
  createLocalServiceDeployValidation,
  createLocalServiceRedeploy,
} from "@rilldata/web-common/runtime-client/local-service";
import { derived, get, writable } from "svelte/store";

export class ProjectDeployer {
  public readonly validation = createLocalServiceDeployValidation({
    query: {
      refetchOnWindowFocus: true,
    },
  });
  public readonly promptOrgSelection = writable(false);

  private readonly deploying = writable(false);
  private readonly deployMutation = createLocalServiceDeploy();
  private readonly redeployMutation = createLocalServiceRedeploy();

  public get isDeployed() {
    const validation = get(this.validation).data as DeployValidationResponse;
    return !!validation?.deployedProjectId;
  }

  public getStatus() {
    return derived(
      [
        this.validation,
        this.deployMutation,
        this.redeployMutation,
        this.deploying,
      ],
      ([validation, deployMutation, redeployMutation, deploying]) => {
        if (
          validation.isFetching ||
          deployMutation.isLoading ||
          redeployMutation.isLoading ||
          deploying
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

  public async validate() {
    let validation = get(this.validation).data as DeployValidationResponse;
    if (validation?.deployedProjectId) {
      return true;
    }

    await waitUntil(() => !get(this.validation).isFetching);
    validation = get(this.validation).data as DeployValidationResponse;

    if (!validation.isAuthenticated) {
      window.open(
        `${validation.loginUrl}/?redirect=${get(page).url.toString()}`,
        "_self",
      );
      return false;
    }

    // Disabling for now. Will support this though "Connect to github"
    // if (
    //   validation.isGithubRepo &&
    //   (!validation.isGithubConnected || !validation.isGithubRepoAccessGranted)
    // ) {
    //   // if the project is a github repo and not connected to github then redirect to grant access
    //   window.open(`${validation.githubGrantAccessUrl}`, "_self");
    //   return false;
    // }

    return true;
  }

  public async deploy(org?: string) {
    // safeguard around deploy
    if (!(await this.validate())) return;

    this.deploying.set(true);
    try {
      const validation = get(this.validation).data as DeployValidationResponse;
      if (validation.deployedProjectId) {
        const resp = await get(this.redeployMutation).mutateAsync({
          projectId: validation.deployedProjectId,
          reupload: !validation.isGithubRepo,
        });
        window.open(resp.frontendUrl, "_self");
      } else {
        if (!org) {
          if (validation.rillUserOrgs.length === 1) {
            org = validation.rillUserOrgs[0];
          } else if (validation.rillUserOrgs.length > 1) {
            this.promptOrgSelection.set(true);
            return;
          } else {
            org = await getOrgName();
          }
        }

        const resp = await get(this.deployMutation).mutateAsync({
          projectName: validation.localProjectName,
          org,
          upload: true, // hardcoded to upload for now
        });
        window.open(resp.frontendUrl + "/-/invite", "_self");
      }
    } catch (err) {
      // no-op
    }
    this.deploying.set(false);
  }
}
