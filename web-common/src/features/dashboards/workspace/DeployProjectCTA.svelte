<script lang="ts" context="module">
  export const allowPrimary = writable(false);
</script>

<script lang="ts">
  import { page } from "$app/stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { getNeverSubscribedIssue } from "@rilldata/web-common/features/billing/issues";
  import TrialDetailsDialog from "@rilldata/web-common/features/billing/TrialDetailsDialog.svelte";
  import ProjectContainsRemoteChangesDialog from "@rilldata/web-common/features/project/ProjectContainsRemoteChangesDialog.svelte";
  import ProjectRedeployConfirmDialog from "@rilldata/web-common/features/project/ProjectRedeployConfirmDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    createLocalServiceGetCurrentProject,
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
    createLocalServiceGitPull,
    createLocalServiceGitStatus,
    createLocalServiceListOrganizationsAndBillingMetadataRequest,
    getLocalServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";
  import Rocket from "svelte-radix/Rocket.svelte";
  import { writable } from "svelte/store";
  import { Button } from "../../../components/button";

  export let hasValidDashboard: boolean;

  let remoteChangeDialog = false;
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

  const gitStatusQuery = createLocalServiceGitStatus();
  $: hasRemoteChanges =
    $gitStatusQuery.data && $gitStatusQuery.data.remoteCommits > 0;
  const gitPullMutation = createLocalServiceGitPull();

  $: ({ isPending: githubPullPending, error: githubPullError } =
    $gitPullMutation);
  let errorFromGitCommand: Error | null = null;
  $: error = githubPullError ?? errorFromGitCommand;

  // gitStatusQuery is refetched. So we have to check `isFetching` to get the correct loading status.
  $: loading = $gitStatusQuery.isFetching || $currentProject.isLoading;

  $: allowPrimary.set(isDeployed || !hasValidDashboard);

  $: user = createLocalServiceGetCurrentUser();
  $: metadata = createLocalServiceGetMetadata();

  $: deployPageUrl = `${$page.url.protocol}//${$page.url.host}/deploy`;

  $: if (userNotLoggedIn && $metadata.data) {
    // Add screen_hint=signup to loginUrl to default to signup page
    const signupUrl = new URL($metadata.data.loginUrl);
    signupUrl.searchParams.set("screen_hint", "signup");
    signupUrl.searchParams.set("redirect", deployPageUrl);
    deployCTAUrl = signupUrl.toString();
  } else {
    deployCTAUrl = deployPageUrl;
  }

  function onRedeploy() {
    if (hasRemoteChanges) {
      remoteChangeDialog = true;
      return;
    }

    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);

    window.open(deployCTAUrl, "_blank");
  }

  function onShowDeploy() {
    if (hasRemoteChanges) {
      remoteChangeDialog = true;
      return;
    }

    if (!isFirstTimeDeploy) {
      // do not show the confirmation dialog for successive deploys
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);
      window.open(deployCTAUrl, "_blank");
      return;
    }

    deployConfirmOpen = true;
    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);
  }

  async function handleForceFetchRemoteCommits() {
    errorFromGitCommand = null;
    const resp = await $gitPullMutation.mutateAsync({
      discardLocal: true,
    });
    // TODO: download diff once API is ready

    void queryClient.invalidateQueries({
      queryKey: getLocalServiceGitStatusQueryKey(),
    });

    if (!resp.output) {
      remoteChangeDialog = false;
      eventBus.emit("notification", {
        message:
          "Remote project changes fetched and merged. Your changes have been stashed.",
      });
      return;
    }

    errorFromGitCommand = new Error(resp.output);
  }
</script>

{#if isDeployed}
  <ProjectRedeployConfirmDialog isLoading={loading} onConfirm={onRedeploy} />
{:else}
  <Tooltip distance={8}>
    <Button
      {loading}
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

<ProjectContainsRemoteChangesDialog
  bind:open={remoteChangeDialog}
  loading={githubPullPending}
  {error}
  onFetchAndMerge={handleForceFetchRemoteCommits}
/>
