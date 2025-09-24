<script lang="ts">
  import {
    createAdminServiceListOrganizationMemberUsersInfinite,
    createAdminServiceUpdateOrganization,
    getAdminServiceGetOrganizationQueryKey,
  } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

  export let open: boolean;
  export let organization: string;
  export let currentBillingContact: string | undefined;

  const PAGE_SIZE = 20;

  $: adminUsersInfinite = createAdminServiceListOrganizationMemberUsersInfinite(
    organization,
    {
      role: OrgUserRoles.Admin,
      pageSize: PAGE_SIZE,
    },
    {
      query: {
        getNextPageParam: (lastPage) => {
          if (lastPage.nextPageToken !== "") {
            return lastPage.nextPageToken;
          }
          return undefined;
        },
      },
    },
  );

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
        message: `${selectedBillingContactLabel} has been assigned as billing contact.`,
      });
    } catch (error) {
      console.error("Error assigning user as billing contact", error);
      eventBus.emit("notification", {
        message:
          "Failed to reassign billing contact. Please try again or contact support.",
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
  <Dialog.Trigger asChild>
    <div class="hidden"></div>
  </Dialog.Trigger>
  <Dialog.Content class="w-[520px]" noClose>
    <Dialog.Header>
      <Dialog.Title>Change billing contact</Dialog.Title>

      <Dialog.Description>
        <div class="mt-2 my-1">
          Select another org admin as billing contact.
        </div>
        <Select
          id="billingContact"
          bind:value={selectedBillingContact}
          options={selectableUsers}
          on:change={({ detail: newName }) =>
            (selectedBillingContact = newName)}
          sameWidth
          fontSize={14}
        />
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer class="mt-3">
      <Button type="secondary" onClick={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        onClick={handleAssignAsBillingContact}
        loading={$updateOrg.isPending}
        disabled={!selectedDifferntBillingContact}
      >
        Assign as billing contact
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
