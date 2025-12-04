<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceUpdateProjectVariables,
    getAdminServiceGetProjectVariablesQueryKey,
  } from "@rilldata/web-admin/client";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog/index.js";
  import { Button } from "@rilldata/web-common/components/button/index.js";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let name: string;
  export let environment: string | undefined;

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  const queryClient = useQueryClient();
  const updateProjectVariables = createAdminServiceUpdateProjectVariables();

  async function onDelete(deletedName: string) {
    try {
      await $updateProjectVariables.mutateAsync({
        org: organization,
        project,
        data: {
          environment,
          unsetVariables: [deletedName],
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

  async function handleDelete() {
    try {
      onDelete(name);
      open = false;
    } catch (error) {
      console.error("Failed to delete environment variable:", error);
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete this environment variable?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          The environment variable <span class="source-code text-sm font-medium"
            >{name}</span
          > will no longer be available for this project.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        onClick={() => {
          open = false;
        }}
      >
        Cancel
      </Button>
      <Button type="primary" status="error" onClick={handleDelete}>
        Yes, delete
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>

<style lang="postcss">
  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
