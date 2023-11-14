<script lang="ts">
  import { createAdminServiceListProjectMembers } from "@rilldata/web-admin/client";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  const membersQuery = createAdminServiceListProjectMembers(
    organization,
    project
  );
  $: owner = $membersQuery.data?.members.find(
    (member) => member.userId === ownerId
  );
</script>

<span
  >{owner?.userName
    ? `Created by ${owner.userName}`
    : "Created through code"}</span
>
