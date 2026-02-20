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
  import type { AxiosError } from "axios";

  export let organization: string;
  export let project: string;

  const updateProjectMutation = createAdminServiceUpdateProject();

  $: projectResp = createAdminServiceGetProject(organization, project);
  $: isPublic = $projectResp.data?.project?.public ?? false;

  async function toggleVisibility() {
    const newVisibility = !isPublic;
    try {
      await $updateProjectMutation.mutateAsync({
        org: organization,
        project: project,
        data: {
          public: newVisibility,
        },
      });

      await queryClient.refetchQueries({
        queryKey: getAdminServiceGetProjectQueryKey(organization, project),
      });

      eventBus.emit("notification", {
        message: newVisibility
          ? "Project is now public"
          : "Project is now private",
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message:
          axiosError.response?.data?.message ?? "Failed to update visibility",
        type: "error",
      });
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

  <Button
    slot="action"
    onClick={toggleVisibility}
    type="secondary-destructive"
    loading={$updateProjectMutation.isPending}
  >
    {isPublic ? "Make private" : "Make public"}
  </Button>
</SettingsContainer>
