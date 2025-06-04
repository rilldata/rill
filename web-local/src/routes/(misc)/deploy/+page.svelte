<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import type { ConnectError } from "@connectrpc/connect";
  import CTAMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventAction } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    createLocalServiceGetCurrentProject,
    createLocalServiceGetCurrentUser,
    createLocalServiceGetMetadata,
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
  const project = createLocalServiceGetCurrentProject();

  $: loading = $user.isPending || $metadata.isPending || $project.isPending;
  $: error = $user.error ?? $metadata.error ?? $project.error;

  $: if (!loading && !error) {
    handleDeploy();
  }

  function handleDeploy() {
    // Should not happen if the servers are up. If not, there should be a query error.
    if (!$user.data || !$metadata.data?.loginUrl) {
      error = {
        message: "Failed to fetch login URL.",
      } as ConnectError;
      return;
    }

    if (!$user.data?.user) {
      // User is not logged in, redirect to login url provided from metadata query.
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginStart);
      const u = new URL($metadata.data?.loginUrl);
      // Set the redirect to this page so that deploy resumes after a login
      u.searchParams.set("redirect", get(page).url.toString());
      window.location.href = u.toString();
      return;
    }

    // User is logged in already.

    void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginSuccess);

    // Cloud project doest exist.
    if (!$project.data?.project) {
      if ($user.data.rillUserOrgs?.length) {
        // If the user has at least one org we show the selector.
        // Note: The selector has the option to create a new org, so we show it even when there is only one org.
        void goto(`/deploy/select-org`);
      } else {
        void goto(`/deploy/create-org`);
      }
      return;
    }

    if (
      $project.data.project.githubUrl &&
      !$project.data.project.managedGitId
    ) {
      // we do not support pushing to a project already connected to user managed github
      error = {
        message: `This project has already been connected to a GitHub repo.
Please push changes directly to GitHub and the project in Rill Cloud will automatically be updated.`,
      } as ConnectError;
      return;
    }
    // Cloud project already exists. Run a redeploy
    void goto(
      `/deploy/redeploy?org=${$project.data.project.orgName}&projectId=${$project.data.project.id}`,
    );
  }
</script>

<!-- This seems to be necessary to trigger tanstack query to update the query object -->
<!-- TODO: find a config to avoid this -->
<div class="hidden">
  {$user.isLoading}-{$metadata.isLoading}-{$project.isLoading}
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
