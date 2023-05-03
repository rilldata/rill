<script lang="ts">
  import { createAdminServiceListOrganizations } from "@rilldata/web-admin/client";
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";

  const orgs = createAdminServiceListOrganizations();
  $: firstOrg = $orgs.data?.organizations?.[0];

  $: addUserCommand = `rill user add --org ${
    firstOrg ? firstOrg.name : "organizationName"
  } --role viewer`;
</script>

<span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
  >Share</span
>
<div>Ask your organization’s admin to invite viewers using the Rill CLI.</div>
<div>
  If you are an admin, run this command in the Rill CLI to invite a teammate to
  view this project.
</div>
<CLICommandDisplay command={addUserCommand} />
