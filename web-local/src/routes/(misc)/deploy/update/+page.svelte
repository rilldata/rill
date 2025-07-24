<script lang="ts">
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types.ts";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags.ts";
  import {
    getIsOrgOnTrial,
    getPlanUpgradeUrl,
  } from "@rilldata/web-common/features/organization/utils.ts";
  import { addPosthogSessionIdToUrl } from "@rilldata/web-common/lib/analytics/posthog.ts";
  import { createLocalServiceRedeploy } from "@rilldata/web-common/runtime-client/local-service.ts";
  import DeployError from "@rilldata/web-common/features/project/DeployError.svelte";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  const orgParam = data.org;
  const projectId = data.projectId;

  const redeployMutation = createLocalServiceRedeploy();

  $: ({ legacyArchiveDeploy } = featureFlags);

  $: error = $redeployMutation.error as Error | null;

  $: planUpgradeUrl = getPlanUpgradeUrl(orgParam);
  $: isOrgOnTrial = getIsOrgOnTrial(orgParam);

  void updateProject(projectId);

  async function updateProject(projectId: string) {
    const resp = await $redeployMutation.mutateAsync({
      projectId,
      // If `legacyArchiveDeploy` is enabled, then use the archive route. Else use upload route.
      // This is mainly set to true in E2E tests.
      reupload: !$legacyArchiveDeploy,
      rearchive: $legacyArchiveDeploy,
    });
    const projectUrl = resp.frontendUrl; // https://ui.rilldata.com/<org>/<project>
    const projectUrlWithSessionId = addPosthogSessionIdToUrl(projectUrl);
    window.open(projectUrlWithSessionId, "_self");
  }

  function onRetry() {
    void updateProject(projectId);
  }

  function onBack() {
    history.back();
  }
</script>

{#if $redeployMutation.isPending}
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
