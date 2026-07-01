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
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { organization, project }: { organization: string; project: string } =
    $props();

  const updateProjectMutation = createAdminServiceUpdateProject();

  let projectResp = $derived(
    createAdminServiceGetProject(organization, project),
  );
  let isPublic = $derived($projectResp.data?.project?.public ?? false);

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
          ? m.settings_project_now_public_notification()
          : m.settings_project_now_private_notification(),
      });
    } catch (err) {
      const axiosError = err as AxiosError<RpcStatus>;
      eventBus.emit("notification", {
        message:
          axiosError.response?.data?.message ?? m.settings_visibility_update_failed(),
        type: "error",
      });
    }
  }
</script>

<SettingsContainer title={m.settings_project_visibility_title()}>
  {#if isPublic}
    {m.settings_project_visibility_public()}
  {:else}
    {m.settings_project_visibility_private()}
  {/if}

  {#snippet action()}
    <Button
      onClick={toggleVisibility}
      type="secondary-destructive"
      loading={$updateProjectMutation.isPending}
    >
      {isPublic ? m.settings_make_private_button() : m.settings_make_public_button()}
    </Button>
  {/snippet}
</SettingsContainer>
