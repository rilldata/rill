<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import {
    getIsOrgOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils.ts";
  import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import {
    createLocalServiceGetProjectRequest,
    createLocalServiceRedeploy,
  } from "@rilldata/web-common/runtime-client/local-service.ts";
  import DeployError from "@rilldata/web-common/features/project/deploy/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import type { PageData } from "./$types";

  export let data: PageData;

  const { orgName, projectName, createManagedRepo } = data;

  $: projectQuery = createLocalServiceGetProjectRequest(orgName, projectName);

  const redeployMutation = createLocalServiceRedeploy();

  $: ({ legacyArchiveDeploy } = featureFlags);

  $: planUpgradeUrl = getPlanUpgradeUrl(orgName ?? "");
  $: isOrgOnTrial = getIsOrgOnTrial(orgName ?? "");

  $: error = ($redeployMutation.error as Error | null) || $projectQuery.error;
  $: loading = $redeployMutation.isPending || $projectQuery.isPending;

  async function updateProject() {
    const project = get(projectQuery).data?.project;
    if (!project) return;
    // We always use Redeploy instead of GitPush.
    // GitPush is a simple wrapper around the `git push` command.
    // 1. It won't switch org/project for the case where we deploy to a project in another org with the same name.
    // 2. It won't switch org/project and create a new managed repo when overwriting a different project.
    // 3. Push any changes to .env since it is in .gitignore. Redeploy has explicit handling for this.
    const resp = await $redeployMutation.mutateAsync({
      projectId: project.id,
      // If `legacyArchiveDeploy` is enabled, then use the archive route. Else use upload route.
      // This is mainly set to true in E2E tests.
      reupload: !$legacyArchiveDeploy,
      rearchive: $legacyArchiveDeploy,
      createManagedRepo,
    });
    const projectUrl = resp.frontendUrl; // https://ui.rilldata.com/<org>/<project>
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
    await waitUntil(() => !get(projectQuery).isPending);
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
