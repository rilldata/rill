<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { page } from "$app/stores";
  import {
    createAdminServiceSetOrganizationMemberUserRole,
    getAdminServiceListOrganizationInvitesQueryKey,
    getAdminServiceListOrganizationMemberUsersQueryKey,
    type V1OrganizationMemberUser,
  } from "@rilldata/web-admin/client";
  import {
    getProjectRolesDescriptionMap,
    getProjectRolesOptions,
  } from "@rilldata/web-admin/features/projects/user-management/constants.ts";
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
  $: projectRolesOptions = getProjectRolesOptions();
  $: projectRolesDescriptions = getProjectRolesDescriptionMap();

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
        message: m.users_guest_upgraded({ role }),
      });
    } catch (error) {
      console.error("Error upgrading user role", error);
      eventBus.emit("notification", {
        message: m.users_error_upgrading_role(),
        type: "error",
      });
    }
    open = false;
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Trigger>
    {#snippet child({ props })}
      <div {...props} class="hidden"></div>
    {/snippet}
  </Dialog.Trigger>
  <Dialog.Content
    class="translate-y-[-200px] md:w-[425px] w-[425px]"
    onInteractOutside={(e) => {
      e.preventDefault();
      open = false;
    }}
  >
    <Dialog.Header>
      <Dialog.Title>{m.users_convert_to_member()}</Dialog.Title>
      <div class="text-sm">{m.users_convert_user_to_role({ user: userName, role })}</div>
    </Dialog.Header>
    <Dialog.Description class="flex flex-col gap-y-2">
      <div class="flex flex-row items-center gap-x-2">
        <Select
          id="org-user-role"
          bind:value={role}
          options={projectRolesOptions}
          full
        />
      </div>
      <div>{projectRolesDescriptions[role]}</div>
    </Dialog.Description>
    <Dialog.Footer>
      <Button type="tertiary" onClick={() => (open = false)}>{m.users_cancel()}</Button>
      <Button
        type="primary"
        onClick={handleUpgrade}
        loading={isPending}
        disabled={isPending}
      >
        {m.users_convert()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>
