<script lang="ts">
  import { page } from "$app/stores";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import FaviconSettings from "@rilldata/web-admin/features/organizations/settings/FaviconSettings.svelte";
  import LogoSettings from "@rilldata/web-admin/features/organizations/settings/LogoSettings.svelte";
  import ThumbnailSettings from "@rilldata/web-admin/features/organizations/settings/ThumbnailSettings.svelte";
  import OrgNameSettings from "@rilldata/web-admin/features/organizations/settings/OrgNameSettings.svelte";
  import OrgDomainAllowListSettings from "@rilldata/web-admin/features/organizations/settings/OrgDomainAllowListSettings.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({
    showUpgradeDialog,
    organizationLogoUrl,
    organizationFaviconUrl,
    organizationThumbnailUrl,
  } = data);

  $: organization = $page.params.organization;
</script>

<OrgNameSettings {organization} />
<LogoSettings {organization} {organizationLogoUrl} />
<FaviconSettings {organization} {organizationFaviconUrl} />
<ThumbnailSettings {organization} {organizationThumbnailUrl} />
<OrgDomainAllowListSettings {organization} />
<!-- disabling for now since  there are some open questions around billing -->
<!--  <DeleteOrg {organization} />-->

{#if showUpgradeDialog}
  <StartTeamPlanDialog open {organization} type="base" />
{/if}
