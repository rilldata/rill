<script lang="ts">
  import { page } from "$app/stores";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import OrgNameSettings from "@rilldata/web-admin/features/organizations/settings/OrgNameSettings.svelte";
  import OrgDomainAllowListSettings from "@rilldata/web-admin/features/organizations/settings/OrgDomainAllowListSettings.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ showUpgradeDialog } = data);

  $: organization = $page.params.organization;
</script>

<OrgNameSettings {organization} />
<OrgDomainAllowListSettings {organization} />
<!-- disabling for now since  there are some open questions around billing -->
<!--  <DeleteOrg {organization} />-->

{#if showUpgradeDialog}
  <StartTeamPlanDialog open {organization} type="base" />
{/if}
