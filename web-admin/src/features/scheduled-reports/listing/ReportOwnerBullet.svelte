<script lang="ts">
  import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  const usersQuery = createAdminServiceSearchProjectUsers(
    organization,
    project,
    { emailQuery: "%", pageSize: 1000, pageToken: undefined }
  );
  $: user = $usersQuery.data?.users.find((user) => user.id === ownerId);
</script>

<span
  >{user?.displayName
    ? `Created by ${user.displayName}`
    : "Created through code"}</span
>
