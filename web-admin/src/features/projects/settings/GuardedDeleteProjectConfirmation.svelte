<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button/index.ts";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
  import {
    createAdminServiceDeleteProject,
    getAdminServiceGetProjectQueryKey,
    getAdminServiceListProjectsForOrganizationQueryKey,
  } from "@rilldata/web-admin/client/index.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { goto } from "$app/navigation";

  let {
    organization,
    project,
    open = $bindable(false),
    button = true,
  }: {
    organization: string;
    project: string;
    open: boolean;
    button: boolean;
  } = $props();

  const deleteProjectMutation = createAdminServiceDeleteProject();

  let deleteProjectResult = $derived($deleteProjectMutation);

  async function deleteProject() {
    await $deleteProjectMutation.mutateAsync({
      org: organization,
      project,
    });

    // Clean up cache before navigating to ensure the new page has fresh data
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

    await goto(`/${organization}`);
  }
</script>

<AlertDialogGuardedConfirmation
  bind:open
  title="Delete Project?"
  description={`The project "${project}" will be permanently deleted along with all its dashboards, data, and settings. This action cannot be undone.`}
  confirmText={`delete ${project}`}
  confirmButtonText="Delete"
  confirmButtonType="destructive"
  loading={deleteProjectResult.isPending}
  error={deleteProjectResult.error?.message}
  onConfirm={deleteProject}
>
  {#if button}
    <Button type="destructive">Delete Project</Button>
  {:else}
    <div class="hidden"></div>
  {/if}
</AlertDialogGuardedConfirmation>
