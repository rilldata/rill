<script lang="ts" context="module">
  export const allowPrimary = writable(false);
</script>

<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TrialDetailsDialog from "@rilldata/web-common/features/billing/TrialDetailsDialog.svelte";
  import { getDeployRoute } from "@rilldata/web-common/features/project/deploy/route-utils.ts";
  import UpdateProjectPopup from "@rilldata/web-common/features/project/deploy/UpdateProjectPopup.svelte";
  import { copyWithAdditionalArguments } from "@rilldata/web-common/lib/url-utils";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import ProjectContainsRemoteChangesDialog from "@rilldata/web-common/features/project/ProjectContainsRemoteChangesDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
    createLocalServiceListMatchingProjectsRequest,
    createLocalServiceGitPull,
    createLocalServiceGitStatus,
    getLocalServiceGitStatusQueryKey,
  } from "@rilldata/web-common/runtime-client/local-service";
  import { onMount } from "svelte";
  import Rocket from "svelte-radix/Rocket.svelte";
  import { writable, get, derived } from "svelte/store";
  import { Button } from "../../../components/button";

  export let hasValidDashboard: boolean;

  let remoteChangeDialog = false;
  let deployConfirmOpen = false;
  let updateProjectDropdownOpen = false;

  const userQuery = createLocalServiceGetCurrentUser();
  const metadata = createLocalServiceGetMetadata();
  const matchingProjectsQuery = createLocalServiceListMatchingProjectsRequest();

  const gitStatusQuery = createLocalServiceGitStatus();
  const gitPullMutation = createLocalServiceGitPull();

  $: ({ isPending: githubPullPending, error: githubPullError } =
    $gitPullMutation);
  let errorFromGitCommand: Error | null = null;
  $: error = githubPullError ?? errorFromGitCommand;

  const deploymentState = derived(
    [gitStatusQuery, userQuery, matchingProjectsQuery],
    ([$git, $user, $projects]) => ({
      // gitStatusQuery is refetched. So we have to check `isFetching` to get the correct loading status.
      loading:
        $git.isFetching || ($user.data?.user ? $projects.isLoading : false),
      isDeployed: !!$projects.data?.projects?.length,
      hasRemoteChanges: $git.data && $git.data.remoteCommits > 0,
    }),
  );
  $: ({ loading, isDeployed, hasRemoteChanges } = $deploymentState);

  $: allowPrimary.set(isDeployed || !hasValidDashboard);

  $: deployPageUrl = getDeployRoute($page);
  $: redirectPageUrl = copyWithAdditionalArguments($page.url, {
    deploy: "true",
  });

  async function onDeploy(resumingDeploy = false) {
    await waitUntil(() => !get(deploymentState).loading);
    if (get(deploymentState).hasRemoteChanges) {
      remoteChangeDialog = true;
      return;
    }

    // Check user login

    const userResp = get(userQuery).data;
    if (!userResp?.user) {
      if (resumingDeploy) {
        // Redirect loop breaker.
        // Right now we set `deploy=true` during a login flow to resume deploy intent.
        // So if it is true without a user, then there was an unexpected error somewhere.
        eventBus.emit("notification", {
          type: "error",
          message: "Authentication failed. Please try deploying again.",
        });
        return;
      }
      // Login url is on a separate domain, so use window.open instead of goto.
      window.location.href = `${$metadata.data!.loginUrl}?redirect=${redirectPageUrl}`;
      return;
    }

    // Check matching projects

    await waitUntil(() => !get(matchingProjectsQuery).isLoading);
    const matchingProjects = get(matchingProjectsQuery).data?.projects;
    if (matchingProjects?.length) {
      updateProjectDropdownOpen = true;
      return;
    }

    if (!userResp.rillUserOrgs?.length) {
      // 1st time user. show a modal explaining the trial period.
      deployConfirmOpen = true;
      return;
    }

    // do not show the confirmation dialog for successive deploys
    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.DeployIntent);
    window.open($deployPageUrl, "_blank");
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

  onMount(() => {
    if ($page.url.searchParams.get("deploy") === "true") {
      // If we are resuming deploy, then unset the param from the url.
      // This prevents the user from saving/sharing a url that would open the deploy dropdown.
      void goto(copyWithAdditionalArguments($page.url, {}, { deploy: false }));
      void onDeploy(true);
    }
  });
</script>

{#if isDeployed && !hasRemoteChanges}
  <UpdateProjectPopup
    bind:open={updateProjectDropdownOpen}
    matchingProjects={$matchingProjectsQuery.data?.projects ?? []}
  />
{:else}
  <Tooltip distance={8}>
    <Button
      {loading}
      onClick={() =>
        hasRemoteChanges ? (remoteChangeDialog = true) : onDeploy()}
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

<TrialDetailsDialog bind:open={deployConfirmOpen} />

<ProjectContainsRemoteChangesDialog
  bind:open={remoteChangeDialog}
  loading={githubPullPending}
  {error}
  onFetchAndMerge={handleForceFetchRemoteCommits}
/>
