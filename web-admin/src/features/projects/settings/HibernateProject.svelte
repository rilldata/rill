<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceHibernateProject,
    createAdminServiceRedeployProject,
    getAdminServiceGetProjectQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let organization: string;
  export let project: string;

  let dialogOpen = false;

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

    dialogOpen = false;

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
      <AlertDialog bind:open={dialogOpen}>
        <AlertDialogTrigger asChild let:builder>
          <Button builders={[builder]} type="secondary">
            Hibernate project
          </Button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Hibernate this project?</AlertDialogTitle>
            <AlertDialogDescription>
              The project will be put into hibernation mode. It can be
              reactivated by accessing it again.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <Button type="secondary" onClick={() => (dialogOpen = false)}>
              Cancel
            </Button>
            <Button
              type="destructive"
              onClick={hibernateProject}
              loading={hibernateResult.isPending}
            >
              Hibernate
            </Button>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    {/if}
  </svelte:fragment>
</SettingsContainer>
