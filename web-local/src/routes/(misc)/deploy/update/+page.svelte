<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import {
    getIsOrgOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils";
  import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import {
    createLocalServiceGetCurrentProject,
    createLocalServiceGetProjectRequest,
    createLocalServiceGitPush,
    createLocalServiceRedeploy,
  } from "@rilldata/web-common/runtime-client/local-service";
  import DeployError from "@rilldata/web-common/features/project/deploy/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { onMount } from "svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  const { orgName, projectName, newManagedRepo } = data;

  $: projectQuery = createLocalServiceGetProjectRequest(orgName, projectName);
  const currentProjectQuery = createLocalServiceGetCurrentProject();

  $: project = $projectQuery.data?.project;
  $: currentProject = $currentProjectQuery.data?.project;

  const redeployMutation = createLocalServiceRedeploy();
  const githubPush = createLocalServiceGitPush();

  $: ({ legacyArchiveDeploy } = featureFlags);

  $: planUpgradeUrl = getPlanUpgradeUrl(orgName ?? "");
  $: isOrgOnTrial = getIsOrgOnTrial(orgName ?? "");

  $: error =
    ($redeployMutation.error as Error | null) ||
    $projectQuery.error ||
    $currentProjectQuery.error;
  $: loading =
    $redeployMutation.isPending ||
    $githubPush.isPending ||
    $projectQuery.isPending ||
    $currentProjectQuery.isPending;

  async function updateProject() {
    if (!project) return;
    let projectUrl = project.frontendUrl;

    // GitPush pushes the changes to the currently connected git remote.
    // So in certain cases if we need to explicitly push to another project's git remote we need to use Redeploy.
    const useRedeploy =
      // Case 1: When project is using legacy archive. New managed repo should be created. This happens only in Redeploy.
      !!project.archiveAssetId ||
      // Case 2: When we are overwriting another project we need to create a new managed repo.
      !!newManagedRepo ||
      // Case 3: When we are pushing to another project of with the same name.
      project.id !== currentProject?.id;

    if (useRedeploy) {
      const resp = await $redeployMutation.mutateAsync({
        projectId: project.id,
        // If `legacyArchiveDeploy` is enabled, then use the archive route. Else use upload route.
        // This is mainly set to true in E2E tests.
        reupload: !$legacyArchiveDeploy,
        rearchive: $legacyArchiveDeploy,
        newManagedRepo: Boolean(newManagedRepo),
      });
      projectUrl = resp.frontendUrl; // https://ui.rilldata.com/<org>/<project>
    } else {
      await $githubPush.mutateAsync({});
    }

    const projectUrlWithSessionId = addPosthogSessionIdToUrl(projectUrl);
    window.open(projectUrlWithSessionId, "_self");
  }

  function onRetry() {
    void updateProject();
  }

  function onBack() {
    window.history.back();
  }

  async function maybeUpdateProject() {
    await waitUntil(() => !loading);
    if (error) return;
    void updateProject();
  }

  onMount(() => {
    void maybeUpdateProject();
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
    {onRetry}
    {onBack}
  />
{/if}
