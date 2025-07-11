<script lang="ts" context="module">
  export const allowPrimary = writable(false);
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getNeverSubscribedIssue } from "@rilldata/web-common/features/billing/issues";
  import TrialDetailsDialog from "@rilldata/web-common/features/billing/TrialDetailsDialog.svelte";
  import ProjectRedeployConfirmDialog from "@rilldata/web-common/features/project/ProjectRedeployConfirmDialog.svelte";
  import PushToGitForDeployDialog from "@rilldata/web-common/features/project/PushToGitForDeployDialog.svelte";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    createLocalServiceGetCurrentProject,
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
    createLocalServiceListOrganizationsAndBillingMetadataRequest,
  } from "@rilldata/web-common/runtime-client/local-service";
  import Rocket from "svelte-radix/Rocket.svelte";
  import { get, writable } from "svelte/store";
  import { Button } from "../../../components/button";

  export let hasValidDashboard: boolean;

  let pushThroughGitOpen = false;
  let deployConfirmOpen = false;
  let deployCTAUrl: string;

  $: orgsMetadata =
    createLocalServiceListOrganizationsAndBillingMetadataRequest();
  $: currentProject = createLocalServiceGetCurrentProject({
    query: {
      refetchOnWindowFocus: true,
    },
  });

  $: isDeployed = !!$currentProject.data?.project;
  $: userNotLoggedIn = !$user.data?.user;
  $: everyOrgHasNeverSubscribed = $orgsMetadata.data?.orgs?.every(
    (o) => !!getNeverSubscribedIssue(o.issues),
  );
  $: isFirstTimeDeploy =
    !isDeployed && (userNotLoggedIn || everyOrgHasNeverSubscribed);

  $: allowPrimary.set(isDeployed || !hasValidDashboard);

  $: user = createLocalServiceGetCurrentUser();
  $: metadata = createLocalServiceGetMetadata();

  $: deployPageUrl = `${$page.url.protocol}//${$page.url.host}/deploy`;

  $: if (userNotLoggedIn && $metadata.data) {
    deployCTAUrl = `${$metadata.data.loginUrl}?redirect=${deployPageUrl}`;
  } else {
    deployCTAUrl = deployPageUrl;
  }

  $: managedGit = $currentProject.data?.project?.managedGitId ? true : false;

  async function onRedeploy() {
    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);

    await waitUntil(() => !get(currentProject).isFetching);
    if (get(currentProject).data?.project?.gitRemote && !managedGit) {
      pushThroughGitOpen = true;
      return;
    }

    window.open(deployCTAUrl, "_blank");
  }

  function onShowDeploy() {
    if (!isFirstTimeDeploy) {
      // do not show the confirmation dialog for successive deploys
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);
      window.open(deployCTAUrl, "_blank");
      return;
    }

    deployConfirmOpen = true;
    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);
  }
</script>

{#if isDeployed}
  <ProjectRedeployConfirmDialog
    isLoading={$currentProject.isLoading}
    onConfirm={onRedeploy}
  />
{:else}
  <Tooltip distance={8}>
    <Button
      loading={$currentProject.isLoading}
      onClick={onShowDeploy}
      type={hasValidDashboard ? "primary" : "secondary"}
    >
      <Rocket size="16px" />

      Deploy
    </Button>
    <TooltipContent slot="tooltip-content">
      Deploy this project to Rill Cloud
    </TooltipContent>
  </Tooltip>
{/if}

<TrialDetailsDialog bind:open={deployConfirmOpen} {deployCTAUrl} />

<PushToGitForDeployDialog
  bind:open={pushThroughGitOpen}
  gitRemote={$currentProject.data?.project?.gitRemote ?? ""}
  subpath={$currentProject.data?.project?.subpath ?? ""}
/>
