<script lang="ts">
  import {
    createAdminServiceListOrganizationMemberUsers,
    createAdminServiceUpdateOrganization,
  } from "@rilldata/web-admin/client";
  import * as Dialog from "@rilldata/web-common/components/dialog-v2";
  import { Button } from "@rilldata/web-common/components/button";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import SearchableMenuContent from "@rilldata/web-common/components/searchable-filter-menu/SearchableMenuContent.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let open: boolean;
  export let organization: string;
  export let currentBillingContact: string | undefined;

  $: adminUsers = createAdminServiceListOrganizationMemberUsers(organization, {
    role: "admin",
  });
  $: selectableUsers =
    $adminUsers.data?.members?.map((u) => ({
      value: u.userEmail,
      label: u.userName ? `${u.userName} (${u.userEmail})` : u.userEmail,
    })) ?? [];
  $: selectedBillingContact = currentBillingContact ?? "";

  const updateOrg = createAdminServiceUpdateOrganization();

  async function handleAssignAsBillingContact() {
    try {
      await $updateOrg.mutateAsync({
        name: organization,
        data: {
          billingEmail: selectedBillingContact,
        },
      });

      eventBus.emit("notification", {
        message: `Successfully assigned ${name} as the billing contact`,
      });
    } catch (error) {
      console.error("Error assigning user as billing contact", error);
      eventBus.emit("notification", {
        message:
          "Error: Unable to assign billing contact. Please try again or contact support if the issue persists.",
        type: "error",
      });
    }

    open = false;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger asChild>
    <div class="hidden"></div>
  </Dialog.Trigger>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Update billing contact</Dialog.Title>

      <Dialog.Description>
        <Select
          id="emails"
          label="Repo"
          bind:value={selectedBillingContact}
          options={selectableUsers}
          on:change={({ detail: newName }) =>
            (selectedBillingContact = newName)}
        />
      </Dialog.Description>
    </Dialog.Header>
    <Dialog.Footer class="mt-3">
      <Button type="secondary" on:click={() => (open = false)}>Cancel</Button>
      <Button type="primary" on:click={handleAssignAsBillingContact}>
        Update
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
