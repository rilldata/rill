<script lang="ts">
  import { createAdminServiceListOrganizationMemberUsers } from "@rilldata/web-admin/client";
  import AvatarSquareList from "../../organizations/users/AvatarSquareList.svelte";

  export let organization: string;

  $: listOrganizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);
  $: organizationUsersCount =
    $listOrganizationMemberUsers.data?.members?.length ?? 0;
</script>

<AvatarSquareList name={organization} count={organizationUsersCount}>
  {@html `Everyone from <span class="font-bold">${organization}</span>`}
</AvatarSquareList>
