import { page } from "$app/stores";
import type { ConnectError } from "@connectrpc/connect";
import { extractDeployError } from "@rilldata/web-common/features/project/deploy-errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import type { DeployValidationResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  createLocalServiceDeploy,
  createLocalServiceDeployValidation,
  createLocalServiceRedeploy,
  getLocalServiceGetCurrentUserQueryKey,
  localServiceGetCurrentUser,
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
          error: extractDeployError(
            (validation.error as ConnectError) ??
              (deployMutation.error as ConnectError) ??
              (redeployMutation.error as ConnectError),
          ).message,
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
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginStart);
      window.open(
        `${validation.loginUrl}/?redirect=${get(page).url.toString()}`,
        "_self",
      );
      return false;
    } else {
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginSuccess);
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
        let checkNextOrg = false;
        if (!org) {
          if (validation.rillUserOrgs.length === 1) {
            org = validation.rillUserOrgs[0];
          } else if (validation.rillUserOrgs.length > 1) {
            this.promptOrgSelection.set(true);
            return;
          } else {
            const userResp = await queryClient.fetchQuery({
              queryKey: getLocalServiceGetCurrentUserQueryKey(),
              queryFn: localServiceGetCurrentUser,
            });
            org = userResp.user!.displayName.replace(/ /g, "");
            checkNextOrg = true;
          }
        }

        // hardcoded to upload for now
        const frontendUrl = await this.tryDeployWithOrg(
          org,
          validation.localProjectName,
          true,
          checkNextOrg,
        );
        window.open(frontendUrl + "/-/invite", "_self");
      }
    } catch (err) {
      // no-op
    }
    this.deploying.set(false);
  }

  private async tryDeployWithOrg(
    org: string,
    projectName: string,
    upload: boolean,
    checkNextOrg: boolean,
  ) {
    let i = 0;

    // eslint-disable-next-line no-constant-condition
    while (true) {
      try {
        const resp = await get(this.deployMutation).mutateAsync({
          projectName,
          org: `${org}${i === 0 ? "" : "-" + i}`,
          upload,
        });
        void behaviourEvent?.fireDeployEvent(
          BehaviourEventAction.DeploySuccess,
        );
        return resp.frontendUrl;
      } catch (e) {
        const err = extractDeployError(e);
        if (err.noAccess && checkNextOrg) {
          i++;
        } else {
          throw e;
        }
      }
    }
  }
}
