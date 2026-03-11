<script lang="ts">
  import {
    createAdminServiceStartDeployment,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListDeploymentsQueryKey,
    V1DeploymentStatus,
  } from "@rilldata/web-admin/client";
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

  $: isStopping = status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING;

  const startMutation = createAdminServiceStartDeployment();

  function handleStart() {
    $startMutation.mutate(
      { deploymentId, data: {} },
      {
        onSuccess: () => {
          void Promise.all([
            queryClient.invalidateQueries({
              queryKey: getAdminServiceGetProjectQueryKey(
                organization,
                project,
                branch ? { branch } : undefined,
              ),
            }),
            queryClient.invalidateQueries({
              queryKey: getAdminServiceListDeploymentsQueryKey(
                organization,
                project,
              ),
            }),
          ]);
        },
      },
    );
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if isStopping}
      <div class="h-36">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
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
          on:click={handleStart}
        >
          Start deployment
        </Button>
      {/if}
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
