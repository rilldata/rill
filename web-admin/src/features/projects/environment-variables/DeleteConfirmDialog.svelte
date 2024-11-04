<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceDeleteUsergroup,
    getAdminServiceListOrganizationMemberUsergroupsQueryKey,
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
  export let id: string;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const deleteUserGroup = createAdminServiceDeleteUsergroup();

  // TODO: rename to onDeleteEnvironmentVariable
  async function onDelete(deletedId: string) {
    try {
      await $deleteUserGroup.mutateAsync({
        organization: organization,
        usergroup: deletedId,
      });

      await queryClient.invalidateQueries(
        getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
      );

      eventBus.emit("notification", { message: "User group deleted" });
    } catch (error) {
      console.error("Error deleting user group", error);
      eventBus.emit("notification", {
        message: "Error deleting user group",
        type: "error",
      });
    }
  }

  async function handleDelete() {
    try {
      onDelete(id);
      open = false;
    } catch (error) {
      console.error("Failed to delete user group:", error);
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
          This environment variable will no longer be available to the project.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}
      >
        Cancel
      </Button>
      <Button type="primary" status="error" on:click={handleDelete}>
        Yes, delete
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
