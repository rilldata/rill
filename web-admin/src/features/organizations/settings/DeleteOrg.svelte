<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
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

<SettingsContainer title="Delete this organization">
  <svelte:fragment slot="body">
    Permanently delete this organization and all of its contents from the Rill
    platform. This action is not reversible — please continue with caution.
  </svelte:fragment>

  <AlertDialogGuardedConfirmation
    slot="action"
    title="Delete this organization?"
    description={`The organization "${organization}" will be permanently deleted along with all its projects, data, and settings. This action cannot be undone.`}
    confirmText={`delete ${organization}`}
    confirmButtonText="Delete"
    confirmButtonType="destructive"
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
</SettingsContainer>
