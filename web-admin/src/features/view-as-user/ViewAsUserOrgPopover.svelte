<script lang="ts">
  import * as Command from "@rilldata/web-common/components/command/index.js";
  import {
    createAdminServiceListOrganizationMemberUsers,
    type V1OrganizationMemberUser,
    type V1User,
  } from "../../client";
  import { errorStore } from "../../components/errors/error-store";
  import { setViewAsUser } from "./viewAsUserStore";

  export let organization: string;
  export let onSelectUser: (user: V1User) => void;

  // Fetch all users in the organization
  $: orgUsersQuery = createAdminServiceListOrganizationMemberUsers(
    organization,
    { pageSize: 1000 },
    {
      query: {
        enabled: !!organization,
      },
    },
  );

  $: orgMembers = $orgUsersQuery.data?.members ?? [];

  // Convert V1OrganizationMemberUser to V1User format
  function memberToUser(member: V1OrganizationMemberUser): V1User {
    return {
      id: member.userId,
      email: member.userEmail,
      displayName: member.userName,
      photoUrl: member.userPhotoUrl,
    };
  }

  function handleViewAsUser(member: V1OrganizationMemberUser) {
    const user = memberToUser(member);
    // Org-level view-as: use a placeholder project name since we're at org level
    // The actual project context will be set when navigating to a project
    // For now, we need to pick the first project the user has access to
    // Since this is org-level, we set isOrgLevel=true and sourceProject to empty
    // The sourceProject will be used for the dropdown, but at org level we use org members API
    setViewAsUser(user, "__org_level__", true);
    errorStore.reset();
    onSelectUser(user);
  }
</script>

<div class="px-0.5 pt-1 pb-2 text-[10px] text-fg-secondary text-left">
  Test your <a
    target="_blank"
    href="https://docs.rilldata.com/build/metrics-view/security#rill-cloud"
    >security policies</a
  > by viewing this organization from the perspective of another user.
</div>

<Command.Root>
  <Command.Input placeholder="Search for users" />
  <Command.List>
    <Command.Empty>No users found.</Command.Empty>
    <Command.Group heading="Organization Members">
      {#each orgMembers as member}
        <Command.Item onSelect={() => handleViewAsUser(member)}>
          {member.userEmail}
        </Command.Item>
      {/each}
    </Command.Group>
  </Command.List>
</Command.Root>
