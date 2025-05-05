import { page } from "$app/stores";
import type { ConnectError } from "@connectrpc/connect";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
import {
  GetCurrentProjectResponse,
  GetCurrentUserResponse,
  GetMetadataResponse,
} from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
import {
  createLocalServiceDeploy,
  createLocalServiceGetCurrentProject,
  createLocalServiceGetCurrentUser,
  createLocalServiceGetMetadata,
  createLocalServiceRedeploy,
} from "@rilldata/web-common/runtime-client/local-service";
import { derived, get, writable } from "svelte/store";
import { addPosthogSessionIdToUrl } from "../../lib/analytics/posthog";

export enum ProjectDeployStage {
  Init,
  Invalid,

  CreateNewOrg,
  SelectOrg,
}

export class ProjectDeployer {
  public readonly metadata = createLocalServiceGetMetadata();
  public readonly user = createLocalServiceGetCurrentUser();
  public readonly project = createLocalServiceGetCurrentProject();

  public stage = writable(ProjectDeployStage.Init);

  private readonly deployMutation = createLocalServiceDeploy();
  private readonly redeployMutation = createLocalServiceRedeploy();

  public getStatus() {
    return derived(
      [
        this.metadata,
        this.user,
        this.project,
        this.deployMutation,
        this.redeployMutation,
      ],
      ([metadata, user, project, deployMutation, redeployMutation]) => {
        if (
          metadata.error ||
          user.error ||
          project.error ||
          deployMutation.error ||
          redeployMutation.error
        ) {
          return {
            isLoading: false,
            error:
              (metadata.error as ConnectError) ??
              (user.error as ConnectError) ??
              (project.error as ConnectError) ??
              (deployMutation.error as ConnectError) ??
              (redeployMutation.error as ConnectError),
          };
        }

        // we can have periods where no mutation is firing, so it can flash to non-loading state
        // to avoid this we always have isLoading to true and only set to false on error
        // since after successful deploy the same page is redirected to cloud we wont have an issue of the final loaded state
        return {
          isLoading: true,
          error: undefined,
        };
      },
    );
  }

  public onSelectOrg() {
    this.stage.set(ProjectDeployStage.SelectOrg);
  }

  public async loginOrDeploy() {
    // Wait for user and metadata to load
    await waitUntil(
      () => !get(this.metadata).isLoading && !get(this.user).isLoading,
    );

    // Check login status
    const metadata = get(this.metadata).data as GetMetadataResponse;
    const userResp = get(this.user).data as GetCurrentUserResponse;
    if (!userResp.user) {
      // If user is not logged in then redirect to login url from metadata
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginStart);
      const u = new URL(metadata.loginUrl);
      u.searchParams.set("redirect", get(page).url.toString());
      window.open(u.toString(), "_self");
    } else {
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginSuccess);
    }

    // Check project status

    // Wait for project request to load
    await waitUntil(() => !get(this.project).isLoading);

    const projectResp = get(this.project).data as GetCurrentProjectResponse;

    // Project already exists
    if (projectResp.project) {
      if (projectResp.project.githubUrl) {
        // We do not support pushing to a project already connected to github as of now
        // Deploy page should not even open in this scenario, so this is just a safeguard.
        this.stage.set(ProjectDeployStage.Invalid);
        return;
      }

      return this.redeploy(projectResp.project.id);
    }

    if (userResp.rillUserOrgs?.length) {
      // If the user has at least one org we show the selector.
      // Note: The selector has the option to create a new org, so we show it even when there is only one org.
      this.stage.set(ProjectDeployStage.SelectOrg);
    } else {
      this.stage.set(ProjectDeployStage.CreateNewOrg);
    }
  }

  public async deploy(org: string) {
    const projectResp = get(this.project).data as GetCurrentProjectResponse;

    const resp = await get(this.deployMutation).mutateAsync({
      org,
      projectName: projectResp.localProjectName,
      upload: true,
    });
    // wait for the telemetry to finish since the page will be redirected after a deploy success
    await behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeploySuccess);
    if (!resp.frontendUrl) return;

    // projectUrl: https://ui.rilldata.com/<org>/<project>
    const projectInviteUrl = resp.frontendUrl + "/-/invite";
    const projectInviteUrlWithSessionId =
      addPosthogSessionIdToUrl(projectInviteUrl);
    window.open(projectInviteUrlWithSessionId, "_self");
  }

  private async redeploy(projectId: string) {
    const resp = await get(this.redeployMutation).mutateAsync({
      projectId,
      reupload: true,
    });
    const projectUrl = resp.frontendUrl; // https://ui.rilldata.com/<org>/<project>
    const projectUrlWithSessionId = addPosthogSessionIdToUrl(projectUrl);
    window.open(projectUrlWithSessionId, "_self");
  }
}
