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
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let {
    organization,
    project,
    open = $bindable(false),
    button = true,
  }: {
    organization: string;
    project: string;
    open?: boolean;
    button?: boolean;
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
      message: m.settings_deleted_project_notification(),
    });

    await goto(`/${organization}`);
  }
</script>

<AlertDialogGuardedConfirmation
  bind:open
  title={m.settings_delete_project_confirm_title()}
  description={m.settings_delete_project_confirm_description({ project })}
  confirmText={`delete ${project}`}
  confirmButtonText={m.settings_delete_button()}
  confirmButtonType="destructive"
  loading={deleteProjectResult.isPending}
  error={deleteProjectResult.error?.message}
  onConfirm={deleteProject}
>
  {#if button}
    <Button type="destructive">{m.settings_delete_project_button()}</Button>
  {:else}
    <div class="hidden"></div>
  {/if}
</AlertDialogGuardedConfirmation>
