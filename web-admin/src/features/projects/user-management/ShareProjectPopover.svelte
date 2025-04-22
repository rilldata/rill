<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceListProjectInvites,
    createAdminServiceListProjectMemberUsergroups,
    createAdminServiceListProjectMemberUsers,
    createAdminServiceListUsergroupMemberUsers,
  } from "@rilldata/web-admin/client";
  import CopyInviteLinkButton from "@rilldata/web-admin/features/projects/user-management/CopyInviteLinkButton.svelte";
  import UserInviteForm from "@rilldata/web-admin/features/projects/user-management/UserInviteForm.svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import AutogroupMembersItem from "./AutogroupMembersItem.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import UsergroupItem from "./UsergroupItem.svelte";
  import UserItem from "./UserItem.svelte";

  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";

  export let organization: string;
  export let project: string;
  export let manageProjectMembers: boolean;

  let isHovered = false;
  let open = false;
  let accessDropdownOpen = false;
  let accessType = "everyone"; // "everyone" or "invite-only"

  $: copyLink = `${$page.url.protocol}//${$page.url.host}/${organization}/${project}`;

  $: listProjectMemberUsergroups =
    createAdminServiceListProjectMemberUsergroups(
      organization,
      project,
      undefined,
      {
        query: {
          refetchOnMount: true,
          refetchOnWindowFocus: true,
        },
      },
    );
  $: listProjectMemberUsers = createAdminServiceListProjectMemberUsers(
    organization,
    project,
    undefined,
    {
      query: {
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );
  $: listProjectInvites = createAdminServiceListProjectInvites(
    organization,
    project,
    undefined,
    {
      query: {
        refetchOnMount: true,
        refetchOnWindowFocus: true,
      },
    },
  );

  $: projectMemberUserGroupsList =
    $listProjectMemberUsergroups.data?.members ?? [];
  $: projectMemberUsersList = $listProjectMemberUsers.data?.members ?? [];
  $: projectInvitesList = $listProjectInvites.data?.invites ?? [];

  $: hasRegularUserGroups = projectMemberUserGroupsList.some(
    (group) => !group.groupManaged,
  );
</script>

<Popover bind:open>
  <PopoverTrigger asChild let:builder>
    <Button builders={[builder]} type="secondary" selected={open}>Share</Button>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[520px]" padding="0">
    <div class="flex flex-col p-4">
      <div class="flex flex-row items-center mb-4">
        <div class="text-sm font-medium">Share project: {project}</div>
        <div class="grow"></div>
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
              <AutogroupMembersItem
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
    <div
      class="flex flex-row items-center px-3.5 py-3 border-t border-gray-200"
    >
      {#if manageProjectMembers}
        <DropdownMenu.Root bind:open={accessDropdownOpen}>
          <DropdownMenu.Trigger>
            <Button
              type="secondary"
              class="flex flex-row items- gap-2"
              forcedStyle="min-height: 28px !important; height: 28px !important; border-color: #D1D5DB !important;"
            >
              {#if accessType === "everyone"}
                <div
                  class="h-4 w-4 flex items-center justify-center bg-primary-600 rounded-sm"
                >
                  <span class="text-[10px] text-white font-semibold"
                    >{organization[0].toUpperCase()}</span
                  >
                </div>
              {:else}
                <Lock size="18px" />
              {/if}
              <span class="text-sm font-medium text-gray-900">
                {#if accessType === "everyone"}
                  Everyone at {organization}
                {:else}
                  Invite only
                {/if}
              </span>
              {#if accessDropdownOpen}
                <CaretUpIcon size="12px" color="text-gray-700" />
              {:else}
                <CaretDownIcon size="12px" color="text-gray-700" />
              {/if}
            </Button>
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end">
            <DropdownMenu.Item
              on:click={() => {
                accessType = "invite-only";
                accessDropdownOpen = false;
              }}
              class="flex flex-col items-start py-2 data-[highlighted]:bg-gray-100 {accessType ===
              'invite-only'
                ? 'bg-gray-50'
                : ''}"
            >
              <div class="flex items-start gap-2">
                <Lock size="20px" color="#374151" />
                <span class="text-xs font-medium text-gray-700"
                  >Invite only</span
                >
              </div>
              <div class="flex flex-row items-center gap-2">
                <div class="w-[20px]" />
                <span class="text-[11px] text-gray-500"
                  >Only admins and invited users can access</span
                >
              </div>
            </DropdownMenu.Item>
            <DropdownMenu.Item
              on:click={() => {
                accessType = "everyone";
                accessDropdownOpen = false;
              }}
              class="flex flex-col items-start py-2 data-[highlighted]:bg-gray-100 {accessType ===
              'everyone'
                ? 'bg-gray-50'
                : ''}"
            >
              <div class="flex items-start gap-2">
                <div
                  class="h-5 w-5 flex items-center justify-center bg-primary-600"
                >
                  <span class="text-xs text-white font-semibold"
                    >{organization[0].toUpperCase()}</span
                  >
                </div>
                <span class="text-xs font-medium text-gray-700"
                  >Everyone at {organization}</span
                >
              </div>
              <div class="flex flex-row items-center gap-2">
                <div class="w-[20px]" />
                <span class="text-[11px] text-gray-500"
                  >Org members can access</span
                >
              </div>
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      {:else}
        <a
          href="https://docs.rilldata.com/manage/user-management#how-to-add-a-project-user"
          target="_blank"
          class="text-xs text-primary-600">Learn more about sharing</a
        >
      {/if}
      <div class="grow"></div>
      <CopyInviteLinkButton {copyLink} />
    </div>
  </PopoverContent>
</Popover>
