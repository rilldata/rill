<!-- Type-to-confirm dialog for deleting a user. Owns its mutation and query
     invalidation so the Users page just renders <DeleteUserDialog>. -->
<script lang="ts">
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";
  import { getAdminServiceSearchUsersQueryKey } from "@rilldata/web-admin/client";
  import { createDeleteUserMutation } from "@rilldata/web-admin/features/superuser/users/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let email: string;

  const deleteUser = createDeleteUserMutation();
  const queryClient = useQueryClient();

  let loading = false;
  let error: string | undefined = undefined;

  async function handleConfirm() {
    loading = true;
    error = undefined;
    try {
      await $deleteUser.mutateAsync({ email });
      eventBus.emit("notification", {
        type: "success",
        message: `User ${email} deleted`,
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceSearchUsersQueryKey(),
      });
    } catch (err) {
      error = `Failed to delete user: ${err}`;
      throw err;
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialogGuardedConfirmation
  bind:open
  title="Delete User"
  description={`This will permanently delete the user ${email}. This action cannot be undone.`}
  confirmText={email}
  confirmButtonText="Delete"
  confirmButtonType="destructive"
  {loading}
  {error}
  onConfirm={handleConfirm}
/>
