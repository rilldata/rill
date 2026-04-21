<!-- Confirm dialog for committing quota changes for an org. -->
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
  import { getAdminServiceGetOrganizationQueryKey } from "@rilldata/web-admin/client";
  import { createUpdateOrgQuotasMutation } from "@rilldata/web-admin/features/superuser/quotas/selectors";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let org: string;
  export let quotas: {
    projects?: number;
    deployments?: number;
    slotsTotal?: number;
    slotsPerDeployment?: number;
    outstandingInvites?: number;
    storageLimitBytesPerDeployment?: string;
  };

  const updateQuotas = createUpdateOrgQuotasMutation();
  const queryClient = useQueryClient();

  let loading = false;

  async function handleConfirm() {
    loading = true;
    try {
      await $updateQuotas.mutateAsync({ data: { org, ...quotas } });
      eventBus.emit("notification", {
        type: "success",
        message: `Quotas updated for org: ${org}`,
      });
      await queryClient.invalidateQueries({
        queryKey: getAdminServiceGetOrganizationQueryKey(org),
      });
      open = false;
    } catch (err) {
      eventBus.emit("notification", {
        type: "error",
        message: `Failed to update quotas: ${err}`,
      });
    } finally {
      loading = false;
    }
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Update Quotas</AlertDialogTitle>
      <AlertDialogDescription>
        This will update the resource quotas for "{org}". This change takes
        effect immediately.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}>Cancel</Button>
      <Button type="primary" onClick={handleConfirm} {loading}>Update</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
