<script lang="ts">
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import AvatarListItem from "@rilldata/web-common/components/avatar/AvatarListItem.svelte";
  import ProjectUserGroupSetRole from "./ProjectUserGroupSetRole.svelte";

  export let organization: string;
  export let group: V1MemberUsergroup;
  export let project: string;
  export let manageOrgAdmins: boolean;
  export let manageOrgMembers: boolean;
</script>

<div class="flex flex-row items-center gap-x-2 justify-between cursor-auto">
  <AvatarListItem
    name={group.groupName}
    count={group.usersCount}
    shape="square"
    showManage={manageOrgMembers}
  />
  <ProjectUserGroupSetRole {organization} {group} {project} {manageOrgAdmins} />
</div>
