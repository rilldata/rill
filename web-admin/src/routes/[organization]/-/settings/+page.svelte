<script lang="ts">
  import { page } from "$app/stores";
  import StartTeamPlanDialog from "@rilldata/web-admin/features/billing/plans/StartTeamPlanDialog.svelte";
  import DangerZone from "@rilldata/web-admin/components/danger-zone/DangerZone.svelte";
  import DeleteOrg from "@rilldata/web-admin/features/organizations/settings/DeleteOrg.svelte";
  import { isBrandingRestricted } from "@rilldata/web-admin/features/billing/plans/utils";
  import FaviconSettings from "@rilldata/web-admin/features/organizations/settings/FaviconSettings.svelte";
  import LogoSettings from "@rilldata/web-admin/features/organizations/settings/LogoSettings.svelte";
  import OrgNameSettings from "@rilldata/web-admin/features/organizations/settings/OrgNameSettings.svelte";
  import OrgDomainAllowListSettings from "@rilldata/web-admin/features/organizations/settings/OrgDomainAllowListSettings.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ showUpgradeDialog, organization: organizationObj } = data);

  $: ({
    logoUrl: organizationLogoUrl,
    logoDarkUrl: organizationLogoDarkUrl,
    faviconUrl: organizationFaviconUrl,
  } = organizationObj ?? {});

  $: organization = $page.params.organization;
  $: brandingDisabled = isBrandingRestricted(
    organizationObj?.billingPlanName ?? "",
  );
</script>

<OrgNameSettings {organization} />
<LogoSettings
  {organization}
  {organizationLogoUrl}
  {organizationLogoDarkUrl}
  disabled={brandingDisabled}
/>
<FaviconSettings
  {organization}
  {organizationFaviconUrl}
  disabled={brandingDisabled}
/>
<OrgDomainAllowListSettings {organization} />

<DangerZone>
  <DeleteOrg {organization} />
</DangerZone>

{#if showUpgradeDialog}
  <StartTeamPlanDialog open {organization} type="base" />
{/if}
