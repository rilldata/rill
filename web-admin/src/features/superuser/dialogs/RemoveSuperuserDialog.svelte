<!-- Confirm dialog for removing a user's superuser access. -->
<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import {
    createAdminServiceSetSuperuser,
    getAdminServiceListSuperusersQueryKey,
  } from "@rilldata/web-admin/client";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let email: string;

  const setSuperuser = createAdminServiceSetSuperuser();
  const queryClient = useQueryClient();

  let loading = false;

  async function handleConfirm() {
    loading = true;
    try {
      await $setSuperuser.mutateAsync({
        data: { email, superuser: false },
      });
      eventBus.emit("notification", {
        type: "success",
        message: `${email} removed as superuser`,
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListSuperusersQueryKey(),
      });
      open = false;
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed: ${err}`,
      });
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Remove Superuser</AlertDialogTitle>
      <AlertDialogDescription>
        Remove superuser access for {email}? They will lose access to this
        console.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}>Cancel</Button>
      <Button type="destructive" onClick={handleConfirm} {loading}
        >Remove</Button
      >
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
