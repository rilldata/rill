<script lang="ts">
  import { createAdminServiceListUsergroupMemberUsers } from "@rilldata/web-admin/client";
  import AvatarSquareList from "../../organizations/users/AvatarSquareList.svelte";

  export let organization: string;
  export let usergroup: string;

  $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
    organization,
    usergroup,
  );
  $: userGroupMembersCount =
    $listUsergroupMemberUsers.data?.members?.length ?? 0;
</script>

<AvatarSquareList name={usergroup} count={userGroupMembersCount}>
  {@html `Everyone from <span class="font-bold">${usergroup}</span>`}
</AvatarSquareList>
