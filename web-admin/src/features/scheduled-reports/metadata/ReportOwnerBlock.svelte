<script lang="ts">
  import { createAdminServiceSearchProjectUsers } from "@rilldata/web-admin/client";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  // Get owner's name
  const usersQuery = createAdminServiceSearchProjectUsers(
    organization,
    project,
    { emailQuery: "%", pageSize: 1000, pageToken: undefined }
  );
  $: user =
    $usersQuery.data &&
    $usersQuery.data.users.find((user) => user.id === ownerId);
</script>

{user?.displayName
  ? `Report created by ${user.displayName}`
  : "Report created through code"} •
