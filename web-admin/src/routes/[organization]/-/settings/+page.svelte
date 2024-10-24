<script lang="ts">
  import { page } from "$app/stores";
  import OrgNameSettings from "@rilldata/web-admin/features/organizations/settings/OrgNameSettings.svelte";
  import OrgDomainAllowListSettings from "@rilldata/web-admin/features/organizations/settings/OrgDomainAllowListSettings.svelte";

  $: organization = $page.params.organization;
</script>

<OrgNameSettings {organization} />
<OrgDomainAllowListSettings {organization} />
<!-- disabling for now since  there are some open questions around billing -->
<!--  <DeleteOrg {organization} />-->
