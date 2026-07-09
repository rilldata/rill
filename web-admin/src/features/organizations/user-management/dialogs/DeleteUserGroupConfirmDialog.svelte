<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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

      eventBus.emit("notification", { message: m.groups_deleted() });
    } catch (error) {
      console.error("Error deleting user group", error);
      eventBus.emit("notification", {
        message: m.groups_error_deleting(),
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
  <AlertDialogTrigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>{m.groups_delete_confirm_title()}</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          {m.groups_delete_confirm_desc()}
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>{m.users_cancel()}</Button
      >
      <Button type="destructive" onClick={handleDelete}>{m.groups_yes_delete()}</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
