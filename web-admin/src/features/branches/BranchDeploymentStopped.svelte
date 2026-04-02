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

  export let organization: string;
  export let project: string;
  export let deploymentId: string;
  export let status: V1DeploymentStatus;
  export let canManage: boolean;
  export let branch: string | undefined;
  export let onStarted: (() => void) | undefined = undefined;

  $: isStopping = status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING;

  const startMutation = createAdminServiceStartDeployment();

  function handleStart() {
    $startMutation.mutate(
      { deploymentId, data: {} },
      {
        onSuccess: () => {
          onStarted?.();

          const projectQueryKey = getAdminServiceGetProjectQueryKey(
            organization,
            project,
            branch ? { branch } : undefined,
          );

          // Without this, the invalidation refetch may return the old STOPPED
          // status (race condition), leaving the UI stuck on this page.
          queryClient.setQueryData<V1GetProjectResponse>(
            projectQueryKey,
            (old) => {
              if (!old?.deployment) return old;
              return {
                ...old,
                deployment: {
                  ...old.deployment,
                  status: V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING,
                },
              };
            },
          );

          // Mark stale without immediate refetch; PENDING triggers polling
          // (1–2s) which picks up the real server status.
          void queryClient.invalidateQueries({
            queryKey: projectQueryKey,
            refetchType: "none",
          });

          void invalidateDeployments(organization, project);
        },
      },
    );
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if isStopping}
      <div class="h-16">
        <Spinner status={EntityStatus.Running} size="3rem" duration={725} />
      </div>
      <CtaHeader variant="bold">Deployment is stopping...</CtaHeader>
    {:else}
      <CtaHeader variant="bold">Deployment stopped</CtaHeader>
      <p class="text-sm text-fg-secondary">
        This branch deployment is not running.
      </p>
      {#if canManage}
        <Button
          type="primary"
          loading={$startMutation.isPending}
          loadingCopy="Starting..."
          onClick={handleStart}
        >
          Start deployment
        </Button>
      {/if}
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
