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
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let groupName: string;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const deleteUserGroup = createAdminServiceDeleteUsergroup();

  async function onDelete(deletedUserGroupName: string) {
    try {
      await $deleteUserGroup.mutateAsync({
        org: organization,
        usergroup: deletedUserGroupName,
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationMemberUsergroupsQueryKey(
          organization,
          {
            includeCounts: true,
          },
        ),
      });

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
      onDelete(groupName);
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
      <AlertDialogTitle>Delete this user group?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          This user group will no longer be able to access the organization.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="plain"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" status="error" onClick={handleDelete}
        >Yes, delete</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
