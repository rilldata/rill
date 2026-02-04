<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceUpdateProject,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import type { AxiosError } from "axios";

  export let organization: string;
  export let project: string;

  const updateProjectMutation = createAdminServiceUpdateProject();

  $: projectResp = createAdminServiceGetProject(organization, project);
  $: isPublic = $projectResp.data?.project?.public ?? false;

  let error: string | undefined = undefined;

  async function makePublic() {
    error = undefined;
    try {
      await $updateProjectMutation.mutateAsync({
        org: organization,
        project: project,
        data: {
          public: true,
        },
      });

      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });

      eventBus.emit("notification", {
        message: "Project is now public",
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      error =
        axiosError.response?.data?.message ?? "Failed to update visibility";
    }
  }

  async function makePrivate() {
    error = undefined;
    try {
      await $updateProjectMutation.mutateAsync({
        org: organization,
        project: project,
        data: {
          public: false,
        },
      });

      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });

      eventBus.emit("notification", {
        message: "Project is now private",
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      error =
        axiosError.response?.data?.message ?? "Failed to update visibility";
    }
  }
</script>

<SettingsContainer title="Project Visibility">
  <svelte:fragment slot="body">
    {#if isPublic}
      This project is currently <strong>Public</strong>. Anyone with the URL can
      view this project.
    {:else}
      This project is currently <strong>Private</strong>. Only members of the
      organization can access this project.
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="action">
    {#if isPublic}
      <Button
        onClick={makePrivate}
        type="secondary"
        loading={$updateProjectMutation.isPending}
      >
        Make private
      </Button>
    {:else}
      <AlertDialogGuardedConfirmation
        title="Make this project public?"
        description={`The project "${project}" will be publicly accessible. Anyone with the URL will be able to view this project.`}
        confirmText={`make ${project} public`}
        loading={$updateProjectMutation.isPending}
        {error}
        onConfirm={makePublic}
      >
        <svelte:fragment let:builder>
          <Button builders={[builder]} type="secondary">Make public</Button>
        </svelte:fragment>
      </AlertDialogGuardedConfirmation>
    {/if}
  </svelte:fragment>
</SettingsContainer>
