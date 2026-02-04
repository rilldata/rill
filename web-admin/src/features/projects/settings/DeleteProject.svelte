<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";

  export let organization: string;
  export let project: string;

  const deleteProjectMutation = createAdminServiceDeleteProject();

  $: deleteProjectResult = $deleteProjectMutation;

  async function deleteProject() {
    await $deleteProjectMutation.mutateAsync({
      org: organization,
      project,
    });

    void goto(`/${organization}`);
    queryClient.removeQueries({
      queryKey: getAdminServiceGetProjectQueryKey(organization, project),
    });
    await queryClient.invalidateQueries({
      queryKey:
        getAdminServiceListProjectsForOrganizationQueryKey(organization),
    });
    eventBus.emit("notification", {
      message: "Deleted project",
    });
  }
</script>

<SettingsContainer title="Delete Project">
  <svelte:fragment slot="body">
    Permanently delete this project and all of its contents from the Rill
    platform. This action is not reversible — please continue with caution.
  </svelte:fragment>

  <AlertDialogGuardedConfirmation
    slot="action"
    title="Delete Project?"
    description={`The project "${project}" will be permanently deleted along with all its dashboards, data, and settings. This action cannot be undone.`}
    confirmText={`delete ${project}`}
    confirmButtonText="Delete"
    confirmButtonType="destructive"
    loading={deleteProjectResult.isPending}
    error={deleteProjectResult.error?.message}
    onConfirm={deleteProject}
  >
    <svelte:fragment let:builder>
      <Button builders={[builder]} type="destructive">Delete Project</Button>
    </svelte:fragment>
  </AlertDialogGuardedConfirmation>
</SettingsContainer>
