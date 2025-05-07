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
    if ($user.data?.user) {
      // User is logged in already.
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginSuccess);

      if ($project.data?.project) {
        // Project already exists. Run a redeploy
        void goto(
          `/deploy/redeploy?org=${$project.data.project.orgName}&projectId=${$project.data.project.id}`,
        );
      } else if ($user.data.rillUserOrgs?.length) {
        // If the user has at least one org we show the selector.
        // Note: The selector has the option to create a new org, so we show it even when there is only one org.
        void goto(`/deploy/select-org`);
      } else {
        void goto(`/deploy/create-org`);
      }
    } else if ($metadata.data?.loginUrl) {
      // User is not logged in, redirect to login url provided from metadata query.
      void behaviourEvent?.fireDeployEvent(BehaviourEventAction.LoginStart);
      const u = new URL($metadata.data?.loginUrl);
      // Set the redirect to this page so that deploy resumes after a login
      u.searchParams.set("redirect", get(page).url.toString());
      void goto(u.toString());
    } else {
      // Should not happen if the servers are up. If not, there would be a query error.
      error = {
        message: "Failed to fetch login URL.",
      } as ConnectError;
    }
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
  <CancelCircleInverse size="7rem" className="text-gray-200" />
  <CTAHeader variant="bold">Oops! An error occurred</CTAHeader>
  <CTAMessage>{error.message}</CTAMessage>
{/if}
