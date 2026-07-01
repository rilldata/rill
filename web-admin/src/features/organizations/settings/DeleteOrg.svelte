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
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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
      message: m.settings_deleted_org_notification(),
    });
    void goto(`/`);
  }
</script>

<SettingsContainer title={m.settings_delete_org_title()}>
  {m.settings_delete_org_description()}

  {#snippet action()}
    <AlertDialogGuardedConfirmation
      title={m.settings_delete_org_confirm_title()}
      description={m.settings_delete_org_confirm_description({ organization })}
      confirmText={`delete ${organization}`}
      confirmButtonText={m.settings_delete_button()}
      confirmButtonType="destructive"
      loading={deleteOrgResult.isPending}
      error={deleteOrgResult.error?.message}
      onConfirm={deleteOrg}
    >
      <Button type="destructive" label={m.settings_delete_org_button()}>
        {m.settings_delete_org_button()}
      </Button>
    </AlertDialogGuardedConfirmation>
  {/snippet}
</SettingsContainer>
