<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetDeployment,
    V1DeploymentStatus,
    type V1Deployment,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";
  import Navigation from "@rilldata/web-common/layout/navigation/Navigation.svelte";
  import { workspaceRoutePrefix } from "@rilldata/web-common/features/workspaces/workspace-routing";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onDestroy } from "svelte";
  import { get } from "svelte/store"; // used for synchronous page.params read
  import EditSessionLoading from "@rilldata/web-admin/features/edit-session/EditSessionLoading.svelte";
  import EditSessionTimeoutBanner from "@rilldata/web-admin/features/edit-session/EditSessionTimeoutBanner.svelte";
  import EditSessionToolbar from "@rilldata/web-admin/features/edit-session/EditSessionToolbar.svelte";
  import {
    invalidateDevDeployments,
    useActiveDevDeployment,
    useCreateDevDeployment,
  } from "@rilldata/web-admin/features/edit-session/use-edit-session";

  // Read params synchronously at initialization; they're stable for the
  // lifetime of this layout (navigating away from /-/edit/ destroys it).
  const { organization, project } = get(page).params;

  // Set the workspace route prefix for cloud editing
  $workspaceRoutePrefix = `/${organization}/${project}/-/edit`;

  const activeDevDeployment = useActiveDevDeployment(organization, project);
  const createMutation = useCreateDevDeployment();

  let createAttempted = false;
  let runtimeHost: string | null = null;
  let instanceId: string | null = null;
  let accessToken: string | null = null;
  let credentialsError: string | null = null;

  const getCredentialsMutation = createAdminServiceGetDeployment();

  // When there's no active dev deployment, create one
  $: if (
    !$activeDevDeployment.isLoading &&
    !$activeDevDeployment.data &&
    !createAttempted &&
    !$createMutation.isPending
  ) {
    createAttempted = true;
    handleCreateDevDeployment();
  }

  // When we have an active running deployment, fetch credentials
  $: activeDeployment = $activeDevDeployment.data;
  $: if (
    activeDeployment?.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    activeDeployment.id &&
    !runtimeHost
  ) {
    fetchCredentials(activeDeployment);
  }

  $: deploymentStatus = activeDeployment?.status;
  $: isLoading =
    $activeDevDeployment.isLoading ||
    $createMutation.isPending ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING;

  $: isErrored =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED;

  $: isReady =
    deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    runtimeHost !== null;

  async function handleCreateDevDeployment() {
    try {
      await $createMutation.mutateAsync({
        org: organization,
        project,
        data: {
          environment: "dev",
          editable: true,
        },
      });
      void invalidateDevDeployments(organization, project);
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to start edit session: ${getRpcErrorMessage(err as any)}`,
      });
    }
  }

  async function fetchCredentials(deployment: V1Deployment) {
    try {
      const resp = await $getCredentialsMutation.mutateAsync({
        deploymentId: deployment.id!,
        data: {},
      });
      runtimeHost = resp.runtimeHost ?? null;
      instanceId = resp.instanceId ?? null;
      accessToken = resp.accessToken ?? null;

      // Point runtime at the dev deployment
      if (runtimeHost && instanceId && accessToken) {
        $runtime = {
          host: runtimeHost,
          instanceId,
          jwt: {
            token: accessToken,
            receivedAt: Date.now(),
            authContext: "user",
          },
        };
      }
    } catch (err) {
      credentialsError =
        getRpcErrorMessage(err as any) ??
        "Failed to get deployment credentials";
    }
  }

  onDestroy(() => {
    // Clear runtime so RuntimeProvider's render gate ({#if host && instanceId})
    // blocks until setRuntime sets production values. Without this, the gate
    // passes immediately with stale dev deployment values.
    runtime.set({ host: "", instanceId: "" });
    $workspaceRoutePrefix = "";
  });
</script>

<div class="edit-session">
  {#if isReady && activeDeployment?.id && instanceId && runtimeHost}
    <EditSessionTimeoutBanner sessionStartedAt={activeDeployment.createdOn} />
    <EditSessionToolbar
      {organization}
      {project}
      deploymentId={activeDeployment.id}
      {instanceId}
    />
    <FileAndResourceWatcher
      host={runtimeHost}
      {instanceId}
      errorBody="Lost connection to the editing environment. Try ending the session and starting a new one."
    >
      <div class="flex flex-1 overflow-hidden">
        <Navigation showFooterLinks={false} />
        <section class="flex-1 overflow-hidden">
          <slot />
        </section>
      </div>
    </FileAndResourceWatcher>
  {:else if isErrored}
    <ErrorPage
      statusCode={500}
      header="Edit session failed"
      body={activeDeployment?.statusMessage ||
        "The editing environment encountered an error. Please try again."}
    />
  {:else if credentialsError}
    <ErrorPage
      statusCode={500}
      header="Failed to connect"
      body={credentialsError}
    />
  {:else if isLoading}
    <EditSessionLoading
      status={deploymentStatus}
      statusMessage={activeDeployment?.statusMessage}
    />
  {/if}
</div>

<style lang="postcss">
  .edit-session {
    @apply flex flex-col h-full;
  }
</style>
