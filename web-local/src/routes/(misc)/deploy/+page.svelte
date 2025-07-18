<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
    createLocalServiceListMatchingProjectsRequest,
  } from "@rilldata/web-common/runtime-client/local-service";
  import CTAHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CTANeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { get } from "svelte/store";

  // It would be great if this could be moved to loader function.
  // But these can take a significant time (~2sec)
  // So until sveltekit supports loading state from loader function these will have to be queried here.
  const user = createLocalServiceGetCurrentUser();
  const metadata = createLocalServiceGetMetadata();
  const matchingProjects = createLocalServiceListMatchingProjectsRequest();

  $: loading =
    $user.isPending || $metadata.isPending || $matchingProjects.isPending;
  let error: Error | null = null;
  $: error = $user.error ?? $metadata.error ?? $matchingProjects.error;

  $: if (!loading && !error) {
    void handleDeploy();
  }

  function handleDeploy() {
    // Should not happen if the servers are up. If not, there should be a query error.
    if (!$user.data || !$metadata.data?.loginUrl) {
      error = new Error("Failed to fetch login URL.");
      return;
    }

    // User is not logged in, redirect to login url provided from metadata query.
    if (!$user.data?.user) {
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginStart);
      const u = new URL($metadata.data?.loginUrl);
      // Set the redirect to this page so that deploy resumes after a login
      u.searchParams.set("redirect", get(page).url.toString());
      window.location.href = u.toString();
      return;
    }

    // User is logged in already.

    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginSuccess);

    // No matching cloud project(s) exist.
    if (!$matchingProjects.data?.projects?.length) {
      if ($user.data.rillUserOrgs?.length) {
        // If the user has at least one org we show the selector.
        // Note: The selector has the option to create a new org, so we show it even when there is only one org.
        return goto(`/deploy/select-org`);
      } else {
        return goto(`/deploy/create-org`);
      }
    }

    if ($matchingProjects.data.projects.length === 1) {
      const singleProject = $matchingProjects.data.projects[0];
      // Project already exists. Run a redeploy
      return goto(
        `/deploy/redeploy?org=${singleProject.orgName}&project=${singleProject.name}`,
      );
    } else {
      return goto(`/deploy/matching-projects`);
    }
  }
</script>

<!-- This seems to be necessary to trigger tanstack query to update the query object -->
<!-- TODO: find a config to avoid this -->
<div class="hidden">
  {$user.isLoading}-{$metadata.isLoading}-{$matchingProjects.isLoading}
</div>

{#if loading}
  <div class="h-36">
    <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
  </div>
  <CTAHeader variant="bold">
    Hang tight! We're deploying your project...
  </CTAHeader>
  <CTANeedHelp />
{:else if error}
  <CancelCircleInverse size="7rem" className="text-gray-200" />
  <CTAHeader variant="bold">Oops! An error occurred</CTAHeader>
  <CTAMessage>{error.message}</CTAMessage>
{/if}
