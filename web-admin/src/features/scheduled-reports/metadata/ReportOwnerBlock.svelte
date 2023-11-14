<script lang="ts">
  import { createAdminServiceListProjectMembers } from "@rilldata/web-admin/client";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  // Get owner's name
  const membersQuery = createAdminServiceListProjectMembers(
    organization,
    project
  );
  $: owner =
    $membersQuery.data &&
    $membersQuery.data.members.find((member) => member.userId === ownerId);
</script>

{owner?.userName
  ? `Report created by ${owner.userName}`
  : "Report created through code"} •
