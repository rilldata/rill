<script lang="ts">
  import {
    createAdminServiceGetProject,
    createAdminServiceHibernateProject,
    getAdminServiceGetProjectQueryKey,
    type RpcStatus,
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
  import type { AxiosError } from "axios";

  export let organization: string;
  export let project: string;

  let dialogOpen = false;

  const hibernateProjectMutation = createAdminServiceHibernateProject();

  $: projectResp = createAdminServiceGetProject(organization, project);
  $: isHibernated = !$projectResp.data?.deployment;

  $: hibernateResult = $hibernateProjectMutation;

  async function hibernateProject() {
    try {
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
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message:
          axiosError.response?.data?.message ?? "Failed to hibernate project",
        type: "error",
      });
    }
  }
</script>

<SettingsContainer title="Hibernate Project">
  <svelte:fragment slot="body">
    Put this project into hibernation mode. Hibernated projects are paused and
    do not consume resources. The project can be woken up at any time.
  </svelte:fragment>

  <svelte:fragment slot="action">
    <AlertDialog bind:open={dialogOpen}>
      <AlertDialogTrigger asChild let:builder>
        <Button
          builders={[builder]}
          type="secondary-destructive"
          disabled={isHibernated}
        >
          Hibernate project
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Hibernate this project?</AlertDialogTitle>
          <AlertDialogDescription>
            The project will be paused and will not consume resources. It can be
            woken up at any time.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <Button type="tertiary" onClick={() => (dialogOpen = false)}>
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
  </svelte:fragment>
</SettingsContainer>
