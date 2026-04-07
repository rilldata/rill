<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
  } from "@rilldata/web-admin/client";
  import DeleteConfirmDialog from "@rilldata/web-common/features/resources/DeleteConfirmDialog.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let name: string;
  export let environment: string | undefined;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const queryClient = useQueryClient();
  const updateProjectVariables = createAdminServiceUpdateProjectVariables();

  async function handleDelete() {
    try {
      await $updateProjectVariables.mutateAsync({
        org: organization,
        project,
        data: {
          environment,
          unsetVariables: [name],
        },
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetProjectVariablesQueryKey(
          organization,
          project,
          {
            forAllEnvironments: true,
          },
        ),
      });

      eventBus.emit("notification", {
        message: "Environment variable deleted",
      });
    } catch (error) {
      console.error("Error deleting environment variable", error);
      eventBus.emit("notification", {
        message: "Error deleting environment variable",
        type: "error",
      });
    }
  }
</script>

<DeleteConfirmDialog
  bind:open
  title="Delete this environment variable?"
  description={`The environment variable <span class="font-mono text-sm font-medium">${name}</span> will no longer be available for this project.`}
  onDelete={handleDelete}
/>
