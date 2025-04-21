<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListProjectInvites,
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceListProjectMemberUsers,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-management/CopyInviteLinkButton.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-management/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import type { V1ProjectInvite } from "@rilldata/web-admin/client";
  import OrganizationItem from "./OrganizationItem.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import UsergroupItem from "./UsergroupItem.svelte";
  import UserItem from "./UserItem.svelte";
  import AvatarListItem from "../../organizations/users/AvatarListItem.svelte";

  export let organization: string;
  export let project: string;

  let open = false;

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;

  $: listProjectMemberUsergroups =
    createAdminServiceListProjectMemberUsergroups(organization, project);
  $: listProjectMemberUsers = createAdminServiceListProjectMemberUsers(
    organization,
    project,
  );
  $: listProjectInvites = createAdminServiceListProjectInvites(
    organization,
    project,
  );

  $: projectMemberUserGroupsList =
    $listProjectMemberUsergroups.data?.members ?? [];
  $: projectMemberUsersList = $listProjectMemberUsers.data?.members ?? [];
  $: projectInvitesList = $listProjectInvites.data?.invites ?? [];

  // $: console.log("projectMemberUsersList: ", projectMemberUsersList);
  $: console.log("projectMemberUserGroupsList: ", projectMemberUserGroupsList);
  // $: console.log("projectInvitesList: ", projectInvitesList);

  $: hasRegularUserGroups = projectMemberUserGroupsList.some(
    (group) => !group.groupManaged,
  );
</script>

<Popover bind:open>
  <PopoverTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary" selected={open}>Share</Button>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[520px] p-4">
    <div class="flex flex-col">
      <div class="flex flex-row items-center mb-4">
        <div class="text-sm font-medium">Share project: {project}</div>
        <div class="grow"></div>
        <CopyInviteLinkButton {copyLink} />
      </div>
      <UserInviteForm {organization} {project} />
      <!-- 52 * 8 = 416px -->
      <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[416px]">
        <div class="mt-4">
          <!-- Project Users -->
          {#each projectMemberUsersList as user}
            <UserItem {organization} {project} {user} />
          {/each}
          <!-- Pending Invites -->
          {#each projectInvitesList as user}
            <UserItem {organization} {project} {user} />
          {/each}
          <!-- User Groups -->
          {#if hasRegularUserGroups}
            {#each projectMemberUserGroupsList as group}
              {#if !group.groupManaged}
                <UsergroupItem {organization} {project} {group} />
              {/if}
            {/each}
          {/if}
        </div>
        <div class="mt-2">
          <div class="text-xs text-gray-500 font-semibold uppercase">
            General Access
          </div>
          <!-- NOTE: Only support "autogroup:members" -->
          <!-- https://www.notion.so/rilldata/User-Management-Role-Based-Access-Control-RBAC-Enhancements-8d331b29d9b64d87bca066e06ef87f54?pvs=4#1acba33c8f5780f4903bf16510193dd8 -->
          {#each projectMemberUserGroupsList as group}
            {#if group.groupName === "autogroup:members"}
              <OrganizationItem
                {organization}
                {project}
                {group}
                avatarName={`Everyone at ${organization}`}
              />
            {/if}
          {/each}
        </div>
      </div>
    </div>
  </PopoverContent>
</Popover>
