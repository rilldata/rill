<script lang="ts">
  import {
    createAdminServiceHibernateProject,
    getAdminServiceGetProjectQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";

  export let organization: string;
  export let project: string;

  const hibernateProjectMutation = createAdminServiceHibernateProject();

  $: hibernateResult = $hibernateProjectMutation;

  async function hibernateProject() {
    await $hibernateProjectMutation.mutateAsync({
      org: organization,
      project: project,
    });

    void queryClient.refetchQueries({
      queryKey: getAdminServiceGetProjectQueryKey(organization, project),
    });

    eventBus.emit("notification", {
      message: "Project hibernated",
    });
  }
</script>

<SettingsContainer title="Hibernate project">
  <svelte:fragment slot="body">
    Put this project into hibernation mode. Hibernated projects are paused and
    do not consume resources. The project can be woken up at any time by
    accessing it.
  </svelte:fragment>

  <AlertDialogGuardedConfirmation
    slot="action"
    title="Hibernate this project?"
    description={`The project ${project} will be put into hibernation mode. It can be reactivated by accessing it again.`}
    confirmText={`hibernate ${project}`}
    loading={hibernateResult.isPending}
    error={hibernateResult.error?.message ?? ""}
    onConfirm={hibernateProject}
  >
    <svelte:fragment slot="default" let:builder>
      <Button builders={[builder]} type="secondary-destructive"
        >Hibernate project</Button
      >
    </svelte:fragment>
  </AlertDialogGuardedConfirmation>
</SettingsContainer>
