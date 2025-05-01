import { page } from "$app/stores";
import type { ConnectError } from "@connectrpc/connect";
import { getTrialIssue } from "@rilldata/web-common/features/billing/issues";
import { getPrettyDeployError } from "@rilldata/web-common/features/project/deploy-errors";
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
  createLocalServiceListOrganizationsAndBillingMetadataRequest,
  createLocalServiceRedeploy,
} from "@rilldata/web-common/runtime-client/local-service";
import { derived, get, writable } from "svelte/store";
import { addPosthogSessionIdToUrl } from "../../lib/analytics/posthog";

export enum ProjectDeployStage {
  Init,

  CreateNewOrg,
  SelectOrg,

  FreshDeploy,
  ReDeploy,
  OverwriteProject,
}

export class ProjectDeployer {
  public readonly metadata = createLocalServiceGetMetadata();
  public readonly orgsMetadata =
    createLocalServiceListOrganizationsAndBillingMetadataRequest();
  public readonly user = createLocalServiceGetCurrentUser();
  public readonly project = createLocalServiceGetCurrentProject();

  public stage = writable(ProjectDeployStage.Init);

  // exposes the exact org being used to deploy.
  // this could change based on user's selection or through auto generation based on user's email
  public readonly org = writable("");
  public readonly orgDisplayName = writable<string | undefined>(undefined);

  private readonly deployMutation = createLocalServiceDeploy();
  private readonly redeployMutation = createLocalServiceRedeploy();

  public constructor(
    // use a specific org. org could be set in url params as a callback from upgrading to team plan
    // this marks the deployer to skip prompting for org selection or auto generation
    private readonly useOrg: string,
  ) {
    this.org.set(useOrg);
  }

  public get isDeployed() {
    const projectResp = get(this.project).data as GetCurrentProjectResponse;
    return !!projectResp?.project;
  }

  public getStatus() {
    return derived(
      [
        this.metadata,
        this.orgsMetadata,
        this.user,
        this.project,
        this.org,
        this.deployMutation,
        this.redeployMutation,
      ],
      ([
        metadata,
        orgsMetadata,
        user,
        project,
        org,
        deployMutation,
        redeployMutation,
      ]) => {
        if (
          metadata.error ||
          orgsMetadata.error ||
          user.error ||
          project.error ||
          deployMutation.error ||
          redeployMutation.error
        ) {
          const orgMetadata = orgsMetadata?.data?.orgs.find(
            (om) => om.name === org,
          );
          const onTrial = !!getTrialIssue(orgMetadata?.issues ?? []);
          return {
            isLoading: false,
            error: getPrettyDeployError(
              (metadata.error as ConnectError) ??
                (user.error as ConnectError) ??
                (project.error as ConnectError) ??
                (deployMutation.error as ConnectError) ??
                (redeployMutation.error as ConnectError),
              onTrial,
            ),
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

  public setOrgAndName(org: string, displayName: string | undefined) {
    this.org.set(org);
    this.orgDisplayName.set(displayName);
    void this.deploy();
  }

  public onNewOrg() {
    this.stage.set(ProjectDeployStage.CreateNewOrg);
  }

  public async loginOrDeploy() {
    await waitUntil(
      () => !get(this.metadata).isLoading && !get(this.user).isLoading,
    );

    const metadata = get(this.metadata).data as GetMetadataResponse;
    const userResp = get(this.user).data as GetCurrentUserResponse;
    if (!userResp.user) {
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginStart);
      const u = new URL(metadata.loginUrl);
      u.searchParams.set("redirect", get(page).url.toString());
      window.open(u.toString(), "_self");
    } else {
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginSuccess);
    }

    return this.deploy();
  }

  public async deploy() {
    await waitUntil(() => !get(this.project).isLoading);

    const projectResp = get(this.project).data as GetCurrentProjectResponse;

    // Project already exists
    if (projectResp.project) {
      this.stage.set(ProjectDeployStage.ReDeploy);
      if (projectResp.project.githubUrl) {
        // we do not support pushing to a project already connected to github as of now
        return;
      }

      const resp = await get(this.redeployMutation).mutateAsync({
        projectId: projectResp.project.id,
        reupload: true,
      });
      const projectUrl = resp.frontendUrl; // https://ui.rilldata.com/<org>/<project>
      const projectUrlWithSessionId = addPosthogSessionIdToUrl(projectUrl);
      window.open(projectUrlWithSessionId, "_self");
      return;
    }

    // Project does not yet exist

    const org = get(this.org);
    if (!org) {
      if (get(this.user).data?.rillUserOrgs?.length) {
        this.stage.set(ProjectDeployStage.SelectOrg);
      } else {
        this.stage.set(ProjectDeployStage.CreateNewOrg);
      }
      return;
    }

    this.stage.set(ProjectDeployStage.FreshDeploy);
    console.log("DEPLOY", {
      org,
      newOrgDisplayName: get(this.orgDisplayName),
      projectName: projectResp.localProjectName,
      upload: true,
    });

    // const resp = await get(this.deployMutation).mutateAsync({
    //   org,
    //   newOrgDisplayName: get(this.orgDisplayName),
    //   projectName: projectResp.localProjectName,
    //   upload: true,
    // });
    // // wait for the telemetry to finish since the page will be redirected after a deploy success
    // await behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeploySuccess);
    // if (!resp.frontendUrl) return;
    //
    // // projectUrl: https://ui.rilldata.com/<org>/<project>
    // const projectInviteUrl = resp.frontendUrl + "/-/invite";
    // const projectInviteUrlWithSessionId =
    //   addPosthogSessionIdToUrl(projectInviteUrl);
    // window.open(projectInviteUrlWithSessionId, "_self");
  }
}
