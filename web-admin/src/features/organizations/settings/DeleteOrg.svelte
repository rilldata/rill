<script lang="ts">
  import { goto } from "$app/navigation";
  import {
    createAdminServiceDeleteOrganization,
    createAdminServiceGetCurrentUser,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import { getActiveOrgLocalStorageKey } from "@rilldata/web-admin/features/organizations/active-org/local-storage";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import AlertDialogGuardedConfirmation from "@rilldata/web-common/components/alert-dialog/alert-dialog-guarded-confirmation.svelte";

  let { organization }: { organization: string } = $props();

  const user = createAdminServiceGetCurrentUser();
  const deleteOrgMutation = createAdminServiceDeleteOrganization();

  let deleteOrgResult = $derived($deleteOrgMutation);

  async function deleteOrg() {
    await $deleteOrgMutation.mutateAsync({
      org: organization,
    });

    // Clear the active org from localStorage to prevent redirect loop
    const userId = $user.data?.user?.id;
    if (userId) {
      const activeOrgKey = getActiveOrgLocalStorageKey(userId);
      localStorage.removeItem(activeOrgKey);
    }

    queryClient.removeQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
    eventBus.emit("notification", {
      message: "Deleted organization",
    });
    void goto(`/`);
  }
</script>

<SettingsContainer title="Delete Organization">
  Permanently delete this organization and all of its contents from the Rill
  platform. This action is not reversible — please continue with caution.

  {#snippet action()}
    <AlertDialogGuardedConfirmation
      title="Delete this organization?"
      description={`The organization "${organization}" will be permanently deleted along with all its projects, data, and settings. This action cannot be undone.`}
      confirmText={`delete ${organization}`}
      confirmButtonText="Delete"
      confirmButtonType="destructive"
      loading={deleteOrgResult.isPending}
      error={deleteOrgResult.error?.message}
      onConfirm={deleteOrg}
    >
      <Button type="destructive" label="Delete organization">
        Delete Organization
      </Button>
    </AlertDialogGuardedConfirmation>
  {/snippet}
</SettingsContainer>
