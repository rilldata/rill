import { page } from "$app/stores";
import type { ConnectError } from "@connectrpc/connect";
import { getTrialIssue } from "@rilldata/web-common/features/billing/issues";
import { sanitizeOrgName } from "@rilldata/web-common/features/organization/sanitizeOrgName";
import {
  DeployErrorType,
  getPrettyDeployError,
} from "@rilldata/web-common/features/project/deploy-errors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
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
  getLocalServiceGetCurrentUserQueryKey,
  localServiceGetCurrentUser,
} from "@rilldata/web-common/runtime-client/local-service";
import { derived, get, writable } from "svelte/store";
import { addPosthogSessionIdToUrl } from "../../lib/analytics/posthog";

export class ProjectDeployer {
  public readonly metadata = createLocalServiceGetMetadata();
  public readonly orgsMetadata =
    createLocalServiceListOrganizationsAndBillingMetadataRequest();
  public readonly user = createLocalServiceGetCurrentUser();
  public readonly project = createLocalServiceGetCurrentProject();
  public readonly promptOrgSelection = writable(false);

  // exposes the exact org being used to deploy.
  // this could change based on user's selection or through auto generation based on user's email
  public readonly org = writable("");

  private readonly deployMutation = createLocalServiceDeploy();
  private readonly redeployMutation = createLocalServiceRedeploy();

  public constructor(
    // use a specific org. org could be set in url params as a callback from upgrading to team plan
    // this marks the deployer to skip prompting for org selection or auto generation
    private readonly useOrg: string,
  ) {}

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

  public async deploy(org?: string) {
    await waitUntil(() => !get(this.project).isLoading);

    const projectResp = get(this.project).data as GetCurrentProjectResponse;

    // Project already exists
    if (projectResp.project) {
      if (projectResp.project.githubUrl) {
        // we do not support pushing to a project already connected to github
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

    if (!org && this.useOrg) {
      org = this.useOrg;
    }

    let checkNextOrg = false;
    if (!org) {
      const { org: inferredOrg, checkNextOrg: inferredCheckNextOrg } =
        await this.inferOrg(get(this.user).data?.rillUserOrgs ?? []);
      // no org was inferred. this is because we have prompted the user for an org
      if (!inferredOrg) return;
      org = inferredOrg;
      checkNextOrg = inferredCheckNextOrg;
    }

    const projectUrl = await this.tryDeployWithOrg(
      org,
      projectResp.localProjectName,
      checkNextOrg,
    );
    if (projectUrl) {
      // projectUrl: https://ui.rilldata.com/<org>/<project>
      const projectInviteUrl = projectUrl + "/-/invite";
      const projectInviteUrlWithSessionId =
        addPosthogSessionIdToUrl(projectInviteUrl);
      window.open(projectInviteUrlWithSessionId, "_self");
    }
  }

  private async inferOrg(rillUserOrgs: string[]) {
    let org: string | undefined;
    let checkNextOrg = false;
    if (rillUserOrgs.length === 1) {
      org = rillUserOrgs[0];
    } else if (rillUserOrgs.length > 1) {
      this.promptOrgSelection.set(true);
    } else {
      const userResp = await queryClient.fetchQuery({
        queryKey: getLocalServiceGetCurrentUserQueryKey(),
        queryFn: localServiceGetCurrentUser,
      });
      org = this.getOrgNameFromEmail(userResp.user?.email ?? "");
      checkNextOrg = true;
    }
    return { org, checkNextOrg };
  }

  private async tryDeployWithOrg(
    org: string,
    projectName: string,
    checkNextOrg: boolean,
  ) {
    let i = 0;

    // eslint-disable-next-line no-constant-condition
    while (true) {
      try {
        const tryOrgName = `${org}${i === 0 ? "" : "-" + i}`;
        this.org.set(tryOrgName);
        const resp = await get(this.deployMutation).mutateAsync({
          projectName,
          org: tryOrgName,
          upload: true,
        });
        // wait for the telemetry to finish since the page will be redirected after a deploy success
        await behaviourEvent?.fireDeployEvent(
          BehaviourEventAction.DeploySuccess,
        );
        return resp.frontendUrl;
      } catch (e) {
        const err = getPrettyDeployError(e, false);
        if (err.type === DeployErrorType.PermissionDenied && checkNextOrg) {
          i++;
        } else {
          throw e;
        }
      }
    }
  }

  private getOrgNameFromEmail(email: string): string {
    return sanitizeOrgName(email.split("@")[0] ?? "");
  }
}
