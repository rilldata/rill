<script lang="ts" context="module">
  export const allowPrimary = writable(false);
</script>

<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import TrialDetailsDialog from "@rilldata/web-common/features/billing/TrialDetailsDialog.svelte";
  import UpdateProjectPopup from "@rilldata/web-common/features/project/deploy/UpdateProjectPopup.svelte";
  import { copyWithAdditionalArguments } from "@rilldata/web-common/lib/url-utils";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
    createLocalServiceListMatchingProjectsRequest,
  } from "@rilldata/web-common/runtime-client/local-service";
  import Rocket from "svelte-radix/Rocket.svelte";
  import { get, writable } from "svelte/store";
  import { Button } from "../../../components/button";

  export let hasValidDashboard: boolean;

  let deployConfirmOpen = false;
  let updateProjectDropdownOpen = false;

  const userQuery = createLocalServiceGetCurrentUser();
  const metadata = createLocalServiceGetMetadata();
  const matchingProjectsQuery = createLocalServiceListMatchingProjectsRequest();

  $: autoOpenDeploy = $page.url.searchParams.get("deploy") === "true";
  $: if (autoOpenDeploy) {
    void onDeploy();
  }

  $: isDeployed = !!$matchingProjectsQuery.data?.projects?.length;

  $: allowPrimary.set(isDeployed || !hasValidDashboard);

  $: deployPageUrl = `${$page.url.protocol}//${$page.url.host}/deploy`;
  $: redirectPageUrl = copyWithAdditionalArguments($page.url, {
    deploy: "true",
  });

  async function onDeploy() {
    let didAutoDeploy = false;
    if (autoOpenDeploy) {
      autoOpenDeploy = false;
      didAutoDeploy = true;
      // If it was an auto-deploy, then unset the param from the url.
      // This prevents the user from saving/sharing a url that would open the deploy dropdown.
      void goto(copyWithAdditionalArguments($page.url, {}, { deploy: false }));
    }

    // Check user login

    await waitUntil(() => !get(userQuery).isLoading);
    const userResp = get(userQuery).data;
    if (!userResp?.user) {
      if (didAutoDeploy) {
        // Redirect loop breaker.
        // Right now we set auto deploy during a login flow.
        // So if it is true without a user then there was an unexpected error somewhere.
        // TODO: show error
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
    window.open(deployPageUrl, "_blank");
  }
</script>

{#if isDeployed}
  <UpdateProjectPopup
    bind:open={updateProjectDropdownOpen}
    matchingProjects={$matchingProjectsQuery.data?.projects ?? []}
  />
{:else}
  <Tooltip distance={8}>
    <Button
      loading={$matchingProjectsQuery.isLoading}
      onClick={onDeploy}
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
