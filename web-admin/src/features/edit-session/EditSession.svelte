<script lang="ts">
  import {
    createAdminServiceGetDeployment,
    V1DeploymentStatus,
    type V1Deployment,
  } from "@rilldata/web-admin/client";
  import { getRpcErrorMessage } from "@rilldata/web-admin/components/errors/error-utils";
  import ErrorPage from "@rilldata/web-common/components/ErrorPage.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import EditSessionLoading from "./EditSessionLoading.svelte";
  import EditSessionTimeoutBanner from "./EditSessionTimeoutBanner.svelte";
  import EditSessionToolbar from "./EditSessionToolbar.svelte";
  import {
    invalidateDevDeployments,
    useActiveDevDeployment,
    useCreateDevDeployment,
  } from "./use-edit-session";

  export let organization: string;
  export let project: string;

  const activeDevDeployment = useActiveDevDeployment(organization, project);
  const createMutation = useCreateDevDeployment();

  // Track whether we've attempted to create a deployment this session
  let createAttempted = false;

  // Credentials for the active dev deployment
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

  // Build iframe URL: runtime host serves web-local, with auth passed as query params
  $: iframeSrc = runtimeHost
    ? buildIframeSrc(runtimeHost, instanceId, accessToken)
    : null;

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

      // Update the runtime store so runtime client functions (e.g. GitPush) target the dev deployment
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

  function buildIframeSrc(
    host: string,
    instId: string | null,
    token: string | null,
  ): string {
    const url = new URL(host);
    if (instId) url.searchParams.set("instance_id", instId);
    if (token) url.searchParams.set("access_token", token);
    return url.toString();
  }
</script>

<div class="edit-session">
  {#if isReady && activeDeployment?.id && instanceId}
    <EditSessionTimeoutBanner sessionStartedAt={activeDeployment.createdOn} />
    <EditSessionToolbar
      {organization}
      {project}
      deploymentId={activeDeployment.id}
      {instanceId}
    />
    {#if iframeSrc}
      <iframe
        src={iframeSrc}
        title="Rill Developer"
        class="iframe"
        allow="clipboard-read; clipboard-write"
      />
    {/if}
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

  .iframe {
    @apply flex-1 w-full border-none;
  }
</style>
