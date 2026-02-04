<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import DangerZoneItem from "@rilldata/web-admin/features/organizations/settings/DangerZoneItem.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";

  export let organization: string;

  const deleteOrgMutation = createAdminServiceDeleteOrganization();

  $: deleteOrgResult = $deleteOrgMutation;

  async function deleteOrg() {
    await $deleteOrgMutation.mutateAsync({
      org: organization,
    });

    void goto(`/`);
    queryClient.removeQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    eventBus.emit("notification", {
      message: "Deleted organization",
    });
  }
</script>

<DangerZoneItem
  title="Delete this organization"
  description="Once you delete an organization, there is no going back. Please be certain."
>
  <AlertDialogGuardedConfirmation
    slot="action"
    title="Delete this organization?"
    description={`The organization "${organization}" will be permanently deleted along with all its projects, data, and settings. This action cannot be undone.`}
    confirmText={`delete ${organization}`}
    loading={deleteOrgResult.isPending}
    error={deleteOrgResult.error?.message}
    onConfirm={deleteOrg}
  >
    <svelte:fragment let:builder>
      <Button builders={[builder]} type="destructive">
        Delete this organization
      </Button>
    </svelte:fragment>
  </AlertDialogGuardedConfirmation>
</DangerZoneItem>
