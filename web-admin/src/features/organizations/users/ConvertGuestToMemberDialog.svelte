<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceSetOrganizationMemberUserRole,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
  } from "@rilldata/web-admin/client";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-management/UserRoleSelect.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let email: string;
  export let isSuperUser: boolean;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();

  let role: string = OrgUserRoles.Viewer;

  async function handleUpgrade() {
    try {
      await $setOrganizationMemberUserRole.mutateAsync({
        organization: organization,
        email: email,
        data: {
          role: role,
        },
      });

      await queryClient.invalidateQueries({
        queryKey:
          getAdminServiceListOrganizationMemberUsersQueryKey(organization),
      });

      await queryClient.invalidateQueries({
        queryKey: getAdminServiceListOrganizationInvitesQueryKey(organization),
      });

      eventBus.emit("notification", {
        message: `Guest upgraded to ${role}`,
      });
    } catch (error) {
      console.error("Error upgrading user role", error);
      eventBus.emit("notification", {
        message: "Error upgrading user role",
        type: "error",
      });
    }
  }
</script>

<Dialog.Root
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
    email = "";
    isSuperUser = false;
  }}
  onOpenChange={(open) => {
    if (!open) {
      email = "";
      isSuperUser = false;
    }
  }}
>
  <Dialog.Trigger asChild>
    <div class="hidden"></div>
  </Dialog.Trigger>
  <Dialog.Content class="translate-y-[-200px]">
    <Dialog.Header>
      <Dialog.Title>Convert {email} to a member</Dialog.Title>
    </Dialog.Header>
    <Dialog.Description class="flex flex-col gap-y-2">
      <div class="flex flex-row items-center gap-x-2">
        <InputLabel label="New Role:" id="role" />
        <UserRoleSelect bind:value={role} />
      </div>
      <div>
        Upgrading a guest to {role} will grant this user access to all open projects
        in the organization. Would you like to upgrade this guest user to {role}?
      </div>
    </Dialog.Description>
    <Dialog.Footer>
      <Button
        type="plain"
        onClick={() => {
          open = false;
        }}
      >
        Cancel
      </Button>
      <Button
        type="primary"
        onClick={() => {
          open = false;
          void handleUpgrade();
        }}
      >
        Change billing contact
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
