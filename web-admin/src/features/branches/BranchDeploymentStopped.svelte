<script lang="ts">
  import {
    createAdminServiceStartDeployment,
    getAdminServiceGetProjectQueryKey,
    V1DeploymentStatus,
    type V1GetProjectResponse,
  } from "@rilldata/web-admin/client";
  import { invalidateDeployments } from "./deployment-utils";
  import { Button } from "@rilldata/web-common/components/button";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let project: string;
  export let deploymentId: string;
  export let status: V1DeploymentStatus;
  export let canManage: boolean;
  export let branch: string | undefined;
  export let starting: boolean = false;

  $: isStopping = status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING;

  const startMutation = createAdminServiceStartDeployment();

  async function handleStart() {
    starting = true;

    try {
      await $startMutation.mutateAsync({ deploymentId, data: {} });

      const projectQueryKey = getAdminServiceGetProjectQueryKey(
        organization,
        project,
        branch ? { branch } : undefined,
      );

      // Without this, the invalidation refetch may return the old STOPPED
      // status (race condition), leaving the UI stuck on this page.
      queryClient.setQueryData<V1GetProjectResponse>(projectQueryKey, (old) => {
        if (!old?.deployment) return old;
        return {
          ...old,
          deployment: {
            ...old.deployment,
            status: V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
          },
        };
      });

      // Mark stale without immediate refetch; PENDING triggers polling
      // (1–2s) which picks up the real server status.
      void queryClient.invalidateQueries({
        queryKey: projectQueryKey,
        refetchType: "none",
      });

      void invalidateDeployments(organization, project);
    } catch (e) {
      console.error("Failed to start deployment", e);
    }
    starting = false;
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if isStopping}
      <div class="h-16">
        <Spinner status={EntityStatus.Running} size="3rem" duration={725} />
      </div>
      <CtaHeader variant="bold">{m.project_hibernating()}</CtaHeader>
    {:else}
      <CtaHeader variant="bold">{m.project_branch_hibernated()}</CtaHeader>
      <p class="text-sm text-fg-secondary">{m.project_branch_is_hibernated()}</p>
      {#if canManage}
        <Button
          type="primary"
          loading={$startMutation.isPending}
          loadingCopy={m.project_starting()}
          onClick={handleStart}
        >
          {m.project_resume_branch()}
        </Button>
      {/if}
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
