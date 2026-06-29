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
  import * as m from "@rilldata/web-common/paraglide/messages.js";

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
        message: m.env_variable_deleted_notification(),
      });
    } catch (error) {
      console.error("Error deleting environment variable", error);
      eventBus.emit("notification", {
        message: m.env_variable_delete_error_notification(),
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
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{m.env_delete_title()}</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {m.env_delete_description({ name })}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}
      >
        {m.env_cancel_button()}
      </Button>
      <Button type="destructive" onClick={handleDelete}>{m.env_yes_delete_button()}</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>

<style lang="postcss">
  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
