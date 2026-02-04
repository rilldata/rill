<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceHibernateProject,
    createAdminServiceRedeployProject,
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
  const redeployProjectMutation = createAdminServiceRedeployProject();

  $: projectResp = createAdminServiceGetProject(organization, project);
  // Project is hibernated if there's no deployment
  $: isHibernated = !$projectResp.data?.deployment;

  $: hibernateResult = $hibernateProjectMutation;
  $: redeployResult = $redeployProjectMutation;

  async function hibernateProject() {
    await $hibernateProjectMutation.mutateAsync({
      org: organization,
      project: project,
    });

    await queryClient.refetchQueries({
      queryKey: getAdminServiceGetProjectQueryKey(organization, project),
    });

    eventBus.emit("notification", {
      message: "Project hibernated",
    });
  }

  async function wakeProject() {
    await $redeployProjectMutation.mutateAsync({
      org: organization,
      project: project,
      data: {},
    });

    await queryClient.refetchQueries({
      queryKey: getAdminServiceGetProjectQueryKey(organization, project),
    });

    eventBus.emit("notification", {
      message: "Project is waking up",
    });
  }
</script>

<SettingsContainer title="Hibernate Project">
  <svelte:fragment slot="body">
    {#if isHibernated}
      This project is currently hibernated. To access, please wake the project.
    {:else}
      Put this project into hibernation mode. Hibernated projects are paused and
      do not consume resources. The project can be woken up at any time.
    {/if}
  </svelte:fragment>

  <svelte:fragment slot="action">
    {#if isHibernated}
      <Button
        onClick={wakeProject}
        type="secondary"
        loading={redeployResult.isPending}
      >
        Wake project
      </Button>
    {:else}
      <AlertDialogGuardedConfirmation
        title="Hibernate this project?"
        description={`The project "${project}" will be put into hibernation mode. It can be reactivated by accessing it again.`}
        confirmText={`hibernate ${project}`}
        loading={hibernateResult.isPending}
        error={hibernateResult.error?.message}
        onConfirm={hibernateProject}
      >
        <svelte:fragment let:builder>
          <Button builders={[builder]} type="secondary"
            >Hibernate project</Button
          >
        </svelte:fragment>
      </AlertDialogGuardedConfirmation>
    {/if}
  </svelte:fragment>
</SettingsContainer>
