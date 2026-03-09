<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceDeleteService,
    getAdminServiceListServicesQueryKey,
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
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let name: string;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const deleteService = createAdminServiceDeleteService();

  async function handleDelete() {
    try {
      await $deleteService.mutateAsync({ org: organization, name });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListServicesQueryKey(organization),
      });

      eventBus.emit("notification", {
        message: `Service "${name}" deleted`,
      });

      open = false;
    } catch (error) {
      console.error("Error deleting service", error);
      eventBus.emit("notification", {
        message: "Error deleting service",
        type: "error",
      });
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild>
    <div class="hidden"></div>
  </AlertDialogTrigger>
  <AlertDialogContent noCancel>
    <AlertDialogHeader>
      <AlertDialogTitle>Delete this service?</AlertDialogTitle>
      <AlertDialogDescription>
        <div class="mt-1">
          The service <span class="font-medium">{name}</span> and all its tokens
          will be permanently deleted.
        </div>
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button
        type="tertiary"
        onClick={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button
        type="destructive"
        onClick={handleDelete}
        disabled={$deleteService.isPending}
      >Yes, delete</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
