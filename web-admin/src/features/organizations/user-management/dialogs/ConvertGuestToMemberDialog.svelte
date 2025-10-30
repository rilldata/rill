<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceSetOrganizationMemberUserRole,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    type V1OrganizationMemberUser,
  } from "@rilldata/web-admin/client";
  import {
    PROJECT_ROLES_DESCRIPTION_MAP,
    PROJECT_ROLES_OPTIONS,
  } from "@rilldata/web-admin/features/projects/constants.ts";
  import { Button } from "@rilldata/web-common/components/button";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let open = false;
  export let user: V1OrganizationMemberUser | undefined;

  $: organization = $page.params.organization;

  const queryClient = useQueryClient();
  const setOrganizationMemberUserRole =
    createAdminServiceSetOrganizationMemberUserRole();
  $: ({ isPending } = $setOrganizationMemberUserRole);

  let role: string = OrgUserRoles.Viewer;

  $: userName = user?.userName ?? user?.userEmail ?? "";

  async function handleUpgrade() {
    if (!user?.userEmail) return;
    try {
      await $setOrganizationMemberUserRole.mutateAsync({
        org: organization,
        email: user.userEmail,
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
        message: `Guest upgraded to member and assigned ${role} role`,
      });
    } catch (error) {
      console.error("Error upgrading user role", error);
      eventBus.emit("notification", {
        message: "Error upgrading user role",
        type: "error",
      });
    }
    open = false;
  }
</script>

<Dialog.Root
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
  }}
>
  <Dialog.Trigger asChild>
    <div class="hidden"></div>
  </Dialog.Trigger>
  <Dialog.Content class="translate-y-[-200px] md:w-[425px] w-[425px]">
    <Dialog.Header>
      <Dialog.Title>Convert to member</Dialog.Title>
      <div class="text-sm">Convert {userName} to {role}</div>
    </Dialog.Header>
    <Dialog.Description class="flex flex-col gap-y-2">
      <div class="flex flex-row items-center gap-x-2">
        <Select
          id="org-user-role"
          bind:value={role}
          options={PROJECT_ROLES_OPTIONS}
          full
        />
      </div>
      <div>{PROJECT_ROLES_DESCRIPTION_MAP[role]}</div>
    </Dialog.Description>
    <Dialog.Footer>
      <Button type="plain" onClick={() => (open = false)}>Cancel</Button>
      <Button
        type="primary"
        onClick={handleUpgrade}
        loading={isPending}
        disabled={isPending}
      >
        Convert
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
