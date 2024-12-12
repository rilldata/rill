<script lang="ts">
  import { createAdminServiceListUsergroupMemberUsers } from "@rilldata/web-admin/client";
  import AvatarListItem from "../../organizations/users/AvatarListItem.svelte";

  export let organization: string;
  export let usergroup: string;

  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    usergroup,
  );
  $: userGroupMembersCount =
    $listUsergroupMemberUsers.data?.members?.length ?? 0;
</script>

<AvatarListItem shape="square" name={usergroup} count={userGroupMembersCount} />
