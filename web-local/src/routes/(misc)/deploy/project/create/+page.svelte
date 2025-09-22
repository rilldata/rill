<script lang="ts">
  import type { ConnectError } from "@connectrpc/connect";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import {
    getIsOrgOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils.ts";
  import { getDeployLandingPage } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import { GithubRepoNoAccessError } from "@rilldata/web-common/features/project/deploy/deploy-errors.ts";
  import { getLocalGitRepoStatus } from "@rilldata/web-common/features/project/selectors.ts";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import {
    createLocalServiceDeploy,
    createLocalServiceGetCurrentProject,
    createLocalServiceGitStatus,
  } from "@rilldata/web-common/runtime-client/local-service.ts";
  import DeployError from "@rilldata/web-common/features/project/deploy/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { onMount } from "svelte";
  import { derived, get } from "svelte/store";
  import type { PageData } from "./$types";

  export let data: PageData;
  const { org: orgParam, useGit } = data;

  const projectQuery = createLocalServiceGetCurrentProject();
  const deployMutation = createLocalServiceDeploy();
  const gitStatusQuery = createLocalServiceGitStatus();
  const gitRepoStatusQuery = getLocalGitRepoStatus();

  $: ({ legacyArchiveDeploy } = featureFlags);

  $: deploymentState = derived(
    [gitStatusQuery, gitRepoStatusQuery, projectQuery, deployMutation],
    ([$git, $gitRepo, $project, $deploy]) => {
      const hasGitUrl = !!$git.data?.githubUrl && !$git.data?.managedGit;
      // Without these variables queries are never fired.
      const gitIsPending = $git.isPending;
      const gitRepoIsPending = $gitRepo.isPending;
      const projectIsPending = $project.isPending;
      const deployIsPending = $deploy.isPending;

      return {
        loading:
          gitIsPending ||
          (hasGitUrl ? gitRepoIsPending : false) ||
          projectIsPending ||
          deployIsPending,
        // TODO: use all git errors except "no repo"
        error: ((hasGitUrl ? $gitRepo.error : undefined) ||
          $project.error ||
          $deploy.error) as ConnectError | undefined,
        githubAccessUrl: $gitRepo.data?.grantAccessUrl ?? "",
      };
    },
  );
  $: ({ loading, error, githubAccessUrl } = $deploymentState);

  $: planUpgradeUrl = getPlanUpgradeUrl(orgParam);
  $: isOrgOnTrial = getIsOrgOnTrial(orgParam);

  async function newProject() {
    const projectResp = get(projectQuery).data;
    const gitRepoStatus = get(gitRepoStatusQuery).data;
    if (!projectResp) return;

    if (useGit && !gitRepoStatus?.hasAccess) {
      // Use an error that DeployError component understands
      error = {
        message: GithubRepoNoAccessError,
      } as ConnectError;
      return;
    }

    const resp = await $deployMutation.mutateAsync({
      org: orgParam,
      projectName: projectResp.localProjectName,
      // If `legacyArchiveDeploy` is enabled, then use the archive route. Else use upload route.
      // This is mainly set to true in E2E tests.
      upload: !$legacyArchiveDeploy && !useGit,
      archive: $legacyArchiveDeploy && !useGit,
    });
    // wait for the telemetry to finish since the page will be redirected after a deploy success
    await behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeploySuccess);
    if (!resp.frontendUrl) return;

    // projectUrl: https://ui.rilldata.com/<org>/<project>
    const projectInviteUrl = getDeployLandingPage(resp.frontendUrl);
    window.open(projectInviteUrl, "_self");
  }

  function onRetry() {
    void newProject();
  }

  function onBack() {
    window.history.back();
  }

  async function maybeNewProject() {
    await waitUntil(() => !loading);
    if (error) return;
    void newProject();
  }

  onMount(() => {
    void maybeNewProject();
  });
</script>

{#if loading}
  <div class="h-36">
    <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
  </div>
  <CTAHeader variant="bold">
    Hang tight! We're deploying your project...
  </CTAHeader>
  <CTANeedHelp />
{:else if error}
  <DeployError
    {error}
    planUpgradeUrl={$planUpgradeUrl}
    isOrgOnTrial={$isOrgOnTrial}
    {githubAccessUrl}
    {onRetry}
    {onBack}
  />
{/if}
