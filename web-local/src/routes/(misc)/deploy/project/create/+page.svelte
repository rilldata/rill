<script lang="ts">
  import { page } from "$app/stores";
  import type { ConnectError } from "@connectrpc/connect";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import {
    getIsOrgOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils.ts";
  import { getGithubAccessUrl } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import {
    GitRepoStatusResponse,
    type GitStatusResponse,
  } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb.ts";
  import {
    createLocalServiceDeploy,
    createLocalServiceGetCurrentProject,
    createLocalServiceGitRepoStatus,
    createLocalServiceGitStatus,
  } from "@rilldata/web-common/runtime-client/local-service.ts";
  import DeployError from "@rilldata/web-common/features/project/deploy/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { onMount } from "svelte";
  import { derived } from "svelte/store";
  import type { PageData } from "./$types";
  import DeployToSelfManagedGitDialog from "@rilldata/web-common/features/project/deploy/DeployToSelfManagedGitDialog.svelte";

  export let data: PageData;
  const { org: orgParam, mode: modeParam } = data;

  const projectQuery = createLocalServiceGetCurrentProject();
  const deployMutation = createLocalServiceDeploy();
  const gitStatusQuery = createLocalServiceGitStatus();
  $: hasGitUrl =
    !!$gitStatusQuery.data?.githubUrl && !$gitStatusQuery.data?.managedGit;
  $: gitRepoStatusQuery = createLocalServiceGitRepoStatus(
    $gitStatusQuery.data?.githubUrl ?? "",
    {
      query: {
        enabled: hasGitUrl,
      },
    },
  );

  $: ({ legacyArchiveDeploy } = featureFlags);

  $: deploymentState = derived(
    [gitStatusQuery, gitRepoStatusQuery, projectQuery, deployMutation],
    ([$git, $gitRepo, $project, $deploy]) => {
      const hasGitUrl = !!$git.data?.githubUrl && !$git.data?.managedGit;

      return {
        loading:
          $git.isPending ||
          (hasGitUrl && $gitRepo.isPending) ||
          $project.isPending ||
          $deploy.isPending,
        error: ($git.error ||
          (hasGitUrl ? $gitRepo.error : undefined) ||
          $project.error ||
          $deploy.error) as ConnectError | undefined,
      };
    },
  );
  $: ({ loading, error } = $deploymentState);
  $: deploying = $deployMutation.isPending;

  $: planUpgradeUrl = getPlanUpgradeUrl(orgParam);
  $: isOrgOnTrial = getIsOrgOnTrial(orgParam);

  let showGithubDeployDialog = false;
  $: ({ githubUrl, branch, subpath } =
    $gitStatusQuery.data ??
    <GitStatusResponse>{ githubUrl: "", branch: "", subpath: "" });

  void newProject(orgParam);
  type Mode = "unknown" | "github" | "rill";
  let lastSeenMode: Mode = (modeParam as Mode | undefined) ?? "unknown";

  async function newProject(orgName: string, mode: Mode = "unknown") {
    if (!$projectQuery.data) return;
    lastSeenMode = mode;
    const projectResp = $projectQuery.data;
    const gitRepoStatus = $gitRepoStatusQuery.data;

    if (mode === "unknown") {
      if (hasGitUrl) {
        showGithubDeployDialog = true;
        return;
      }

      mode = "rill";
    } else if (mode === "github" && !gitRepoStatus?.hasAccess) {
      if (gitRepoStatus) {
        window.location.href = getGithubAccessUrl(
          gitRepoStatus.grantAccessUrl,
          $page.url,
        );
      } else {
        // TODO: more comprehensive errors
      }
    }
    const useRillManaged = mode === "rill";

    const resp = await $deployMutation.mutateAsync({
      org: orgName,
      projectName: projectResp.localProjectName,
      // If `legacyArchiveDeploy` is enabled, then use the archive route. Else use upload route.
      // This is mainly set to true in E2E tests.
      upload: !$legacyArchiveDeploy && useRillManaged,
      archive: $legacyArchiveDeploy && useRillManaged,
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

  function onRetry() {
    void newProject(orgParam, lastSeenMode);
  }

  function onBack() {
    window.history.back();
  }

  async function maybeNewProject() {
    await waitUntil(() => !loading);
    if (error) return;
    void newProject(orgParam);
  }

  onMount(() => {
    void maybeNewProject();
  });
</script>

{#if loading}
  <div class="h-36">
    <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
  </div>
  {#if deploying}
    <CTAHeader variant="bold">
      Hang tight! We're deploying your project...
    </CTAHeader>
    <CTANeedHelp />
  {/if}
{:else if error}
  <DeployError
    {error}
    planUpgradeUrl={$planUpgradeUrl}
    isOrgOnTrial={$isOrgOnTrial}
    {onRetry}
    {onBack}
  />
{/if}

<DeployToSelfManagedGitDialog
  bind:open={showGithubDeployDialog}
  {githubUrl}
  {branch}
  {subpath}
  onUseRill={() => newProject(orgParam, "rill")}
  onUseGithub={() => newProject(orgParam, "github")}
/>
