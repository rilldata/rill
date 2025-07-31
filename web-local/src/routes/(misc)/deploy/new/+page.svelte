<script lang="ts">
  import type { ConnectError } from "@connectrpc/connect";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import {
    getIsOrgOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils.ts";
  import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics.ts";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
  import { GetCurrentProjectResponse } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb.ts";
  import {
    createLocalServiceDeploy,
    createLocalServiceGetCurrentProject,
  } from "@rilldata/web-common/runtime-client/local-service.ts";
  import DeployError from "@rilldata/web-common/features/project/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  const orgParam = data.org;

  const project = createLocalServiceGetCurrentProject();
  const deployMutation = createLocalServiceDeploy();

  $: ({ legacyArchiveDeploy } = featureFlags);

  $: error = $deployMutation.error as ConnectError;

  $: planUpgradeUrl = getPlanUpgradeUrl(orgParam);
  $: isOrgOnTrial = getIsOrgOnTrial(orgParam);

  void newProject(orgParam);

  async function newProject(orgName: string) {
    await waitUntil(() => !!$project.data);
    const projectResp = $project.data as GetCurrentProjectResponse;

    const resp = await $deployMutation.mutateAsync({
      org: orgName,
      projectName: projectResp.localProjectName,
      // If `legacyArchiveDeploy` is enabled, then use the archive route. Else use upload route.
      // This is mainly set to true in E2E tests.
      upload: !$legacyArchiveDeploy,
      archive: $legacyArchiveDeploy,
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
    void newProject(orgParam);
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
    isOrgOnTrial={$isOrgOnTrial}
    {onRetry}
    {onBack}
  />
{/if}
