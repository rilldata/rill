<script lang="ts">
  import { createAdminServiceListProjectMemberUsergroups } from "@rilldata/web-admin/client";
  import CreatePublicURLForm from "@rilldata/web-admin/features/public-urls/CreatePublicURLForm.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Link from "@rilldata/web-common/components/icons/Link.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import {
    Tabs,
    TabsContent,
    TabsList,
    TabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import OrganizationItem from "../../projects/user-management/OrganizationItem.svelte";
  import UsergroupItem from "../../projects/user-management/UsergroupItem.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";

  export let createMagicAuthTokens: boolean;
  export let organization: string;
  export let project: string;

  let isOpen = false;
  let copied = false;

  function onCopy() {
    navigator.clipboard.writeText(window.location.href).catch(console.error);
    copied = true;

    setTimeout(() => {
      copied = false;
    }, 2_000);
  }

  $: listProjectMemberUsergroups =
    createAdminServiceListProjectMemberUsergroups(organization, project);
  $: projectMemberUserGroupsList =
    $listProjectMemberUsergroups.data?.members ?? [];

  $: showOrganizationSection = projectMemberUserGroupsList.some(
    (group) => group.groupName === "all-users",
  );
  $: showAllUsersGroup = projectMemberUserGroupsList.find(
    (group) => group.groupName === "all-users",
  );
  $: showGroupsSection =
    projectMemberUserGroupsList.length > 0 &&
    projectMemberUserGroupsList.length === 1 &&
    projectMemberUserGroupsList[0].groupName !== "all-users";
</script>

<Popover bind:open={isOpen}>
  <PopoverTrigger asChild let:builder>
    <Button type="secondary" builders={[builder]} selected={isOpen}
      >Share</Button
    >
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[402px] p-0">
    <Tabs>
      <TabsList>
        <TabsTrigger value="tab1">Copy URL</TabsTrigger>
        {#if createMagicAuthTokens}
          <TabsTrigger value="tab2">Create public URL</TabsTrigger>
        {/if}
      </TabsList>
      <TabsContent value="tab1" class="mt-0 p-4">
        <div class="flex flex-col gap-y-4">
          <h3 class="text-xs text-gray-800 font-normal">
            Share your current view with another project member.
          </h3>
          <Button
            type="secondary"
            on:click={() => {
              onCopy();
            }}
          >
            {#if copied}
              <Check size="16px" />
              Copied URL
            {:else}
              <Link size="16px" className="text-primary-500" />
              Copy URL for this view
            {/if}
          </Button>
        </div>
        {#if showOrganizationSection}
          <div class="mt-4">
            <div class="text-xs text-gray-500 font-semibold uppercase">
              Organization
            </div>
            <div class="flex flex-col gap-y-1">
              <OrganizationItem
                {organization}
                {project}
                group={showAllUsersGroup ?? null}
                canManage
              />
            </div>
          </div>
        {/if}
        {#if showGroupsSection}
          <div class="mt-2">
            <div class="text-xs text-gray-500 font-semibold uppercase">
              Groups
            </div>
            <!-- 52 * 5 = 260px -->
            <div class="flex flex-col gap-y-1 overflow-y-auto max-h-[260px]">
              {#each projectMemberUserGroupsList as group}
                <UsergroupItem {organization} {project} {group} canManage />
              {/each}
            </div>
          </div>
        {/if}
      </TabsContent>
      <TabsContent value="tab2" class="mt-0 p-4">
        <CreatePublicURLForm />
      </TabsContent>
    </Tabs>
  </PopoverContent>
</Popover>

<style lang="postcss">
  h3 {
    @apply font-semibold;
  }
</style>
