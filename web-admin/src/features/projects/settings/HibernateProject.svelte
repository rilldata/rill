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
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  let { organization, project }: { organization: string; project: string } =
    $props();

  let dialogOpen = $state(false);

  const hibernateProjectMutation = createAdminServiceHibernateProject();

  let projectResp = $derived(
    createAdminServiceGetProject(organization, project),
  );
  let isHibernated = $derived(!$projectResp.data?.deployment);

  let hibernateResult = $derived($hibernateProjectMutation);

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
        message: m.settings_project_hibernated_notification(),
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message:
          axiosError.response?.data?.message ?? m.settings_hibernate_failed_notification(),
        type: "error",
      });
    }
  }
</script>

<SettingsContainer title={m.settings_hibernate_project_title()}>
  {m.settings_hibernate_project_description()}

  {#snippet action()}
    <AlertDialog bind:open={dialogOpen}>
      <AlertDialogTrigger>
        {#snippet child({ props })}
          <Button
            {...props}
            type="secondary-destructive"
            disabled={isHibernated}
          >
            {m.settings_hibernate_project_button()}
          </Button>
        {/snippet}
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{m.settings_hibernate_confirm_title()}</AlertDialogTitle>
          <AlertDialogDescription>
            {m.settings_hibernate_confirm_description()}
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <Button type="tertiary" onClick={() => (dialogOpen = false)}>
            {m.settings_cancel_button()}
          </Button>
          <Button
            type="destructive"
            onClick={hibernateProject}
            loading={hibernateResult.isPending}
          >
            {m.settings_hibernate_button()}
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  {/snippet}
</SettingsContainer>
