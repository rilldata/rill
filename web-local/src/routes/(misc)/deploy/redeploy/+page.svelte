<script lang="ts">
  import { page } from "$app/stores";
  import type { ConnectError } from "@connectrpc/connect";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    getOrgIsOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils";
  import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog";
  import type { Project } from "@rilldata/web-common/proto/gen/rill/admin/v1/api_pb.ts";
  import {
    createLocalServiceGetProjectRequest,
    createLocalServiceGitPush,
    createLocalServiceRedeploy,
  } from "@rilldata/web-common/runtime-client/local-service";
  import DeployError from "@rilldata/web-common/features/project/deploy/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: orgName = $page.url.searchParams.get("org") ?? "";
  $: projectName = $page.url.searchParams.get("project") ?? "";

  $: projectQuery = createLocalServiceGetProjectRequest(orgName, projectName);
  $: project = $projectQuery.data?.project;

  $: if (project) void redeploy(project);

  const redeployMutation = createLocalServiceRedeploy();
  const githubPush = createLocalServiceGitPush();

  $: error = $redeployMutation.error as ConnectError;

  $: planUpgradeUrl = getPlanUpgradeUrl(orgName ?? "");
  $: orgIsOnTrial = getOrgIsOnTrial(orgName ?? "");

  $: loading = $redeployMutation.isPending || $githubPush.isPending;

  async function redeploy(project: Project) {
    let projectUrl = project.frontendUrl;
    if (project.archiveAssetId) {
      // Legacy archive based project. Use redeploy API.
      const resp = await $redeployMutation.mutateAsync({
        projectId: project.id,
        reupload: true,
      });
      projectUrl = resp.frontendUrl; // https://ui.rilldata.com/<org>/<project>
    } else {
      await $githubPush.mutateAsync({});
    }

    const projectUrlWithSessionId = addPosthogSessionIdToUrl(projectUrl);
    window.open(projectUrlWithSessionId, "_self");
  }

  function onRetry() {
    void redeploy(project!);
  }

  function onBack() {
    history.back();
  }
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
    orgIsOnTrial={$orgIsOnTrial}
    {onRetry}
    {onBack}
  />
{/if}
