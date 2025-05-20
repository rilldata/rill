<script lang="ts">
  import { page } from "$app/stores";
  import type { ConnectError } from "@connectrpc/connect";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import {
    getOrgIsOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils";
  import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { GetCurrentProjectResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
  import {
    createLocalServiceDeploy,
    createLocalServiceGetCurrentProject,
  } from "@rilldata/web-common/runtime-client/local-service";
  import DeployError from "@rilldata/web-common/features/project/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: orgParam = $page.url.searchParams.get("org");

  $: if (orgParam) void freshDeploy(orgParam);

  const project = createLocalServiceGetCurrentProject();
  const deployMutation = createLocalServiceDeploy();

  $: error = $deployMutation.error as ConnectError;

  $: planUpgradeUrl = getPlanUpgradeUrl(orgParam ?? "");
  $: orgIsOnTrial = getOrgIsOnTrial(orgParam ?? "");

  async function freshDeploy(orgName: string) {
    const projectResp = $project.data as GetCurrentProjectResponse;

    const resp = await $deployMutation.mutateAsync({
      org: orgName,
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

  function onRetry() {
    void freshDeploy(orgParam!);
  }

  function onBack() {
    history.back();
  }
</script>

{#if $deployMutation.isPending}
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
