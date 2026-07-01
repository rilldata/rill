<script lang="ts">
  import {
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import { getOrgAdminMembers } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";

  export let open: boolean;
  export let organization: string;
  export let currentBillingContact: string | undefined;

  $: adminUsersInfinite = getOrgAdminMembers(organization);

  // Flatten all pages of admin users
  $: allAdminUsers =
    $adminUsersInfinite.data?.pages.flatMap((page) => page.members ?? []) ?? [];

  $: selectableUsers = allAdminUsers.map((u) => ({
    value: u.userEmail,
    label: u.userName ? `${u.userName} (${u.userEmail})` : u.userEmail,
  }));
  $: selectedBillingContact = currentBillingContact ?? "";
  $: selectedDifferntBillingContact =
    currentBillingContact !== selectedBillingContact;

  const updateOrg = createAdminServiceUpdateOrganization();

  async function handleAssignAsBillingContact() {
    const selectedBillingContactUser = allAdminUsers.find(
      (u) => u.userEmail === selectedBillingContact,
    );
    const selectedBillingContactName =
      selectedBillingContactUser?.userName ?? selectedBillingContact;
    const selectedBillingContactLabel = `${selectedBillingContactName} (${selectedBillingContact})`;

    try {
      await $updateOrg.mutateAsync({
        org: organization,
        data: {
          billingEmail: selectedBillingContact,
        },
      });

      eventBus.emit("notification", {
        message: m.billing_contact_assigned({ name: selectedBillingContactLabel }),
      });
    } catch (error) {
      console.error("Error assigning user as billing contact", error);
      eventBus.emit("notification", {
        message: m.billing_contact_reassign_failed(),
        type: "error",
      });
    }

    open = false;
    await queryClient.invalidateQueries({
      queryKey: getAdminServiceGetOrganizationQueryKey(organization),
    });
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Content class="w-[520px]" noClose>
    <Dialog.Header>
      <Dialog.Title>{m.billing_change_billing_contact()}</Dialog.Title>

      <Dialog.Description>
        <div class="mt-2 my-1">
          {m.billing_select_admin_as_contact()}
        </div>
        <Select
          id="billingContact"
          bind:value={selectedBillingContact}
          options={selectableUsers}
          onChange={(newName) => (selectedBillingContact = newName)}
          sameWidth
          fontSize={14}
        />
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer class="mt-3">
      <Button type="secondary" onClick={() => (open = false)}>{m.billing_cancel()}</Button>
      <Button
        type="primary"
        onClick={handleAssignAsBillingContact}
        loading={$updateOrg.isPending}
        disabled={!selectedDifferntBillingContact}
      >
        {m.billing_assign_as_contact()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
