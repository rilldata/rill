<script lang="ts">
  import {
    createAdminServiceGetOrganization,
    createAdminServiceListProjectsForOrganization,
  } from "@rilldata/web-admin/client";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { formatDataSizeQuota } from "@rilldata/web-admin/features/billing/plans/utils";
  import { Progress } from "@rilldata/web-common/components/progress";

  export let organization: string;

  $: projects = createAdminServiceListProjectsForOrganization(organization);
  $: organizationQuotas = createAdminServiceGetOrganization(organization, {
    query: {
      select: (data) => data.organization?.quotas,
    },
  });

  $: projectQuota =
    $organizationQuotas.data?.projects &&
    $organizationQuotas.data?.projects !== -1
      ? $organizationQuotas.data?.projects
      : "Unlimited";

  $: usageMetrics = getOrganizationUsageMetrics(organization);
  $: totalOrgSize = $usageMetrics?.data?.reduce((s, m) => s + m.size, 0) ?? 0;

  $: singleProjectLimit = $organizationQuotas.data?.projects === 1;
  $: storageLimitBytesPerDeployment =
    $organizationQuotas.data?.storageLimitBytesPerDeployment ?? "";
</script>

<div class="quotas">
  <div class="quota-entry">
    <div class="quota-entry-title">Projects</div>
    <div class="quota-entry-body">
      {$projects.data?.projects?.length} of {projectQuota}
    </div>
  </div>

  {#if $usageMetrics?.data}
    {#if singleProjectLimit && storageLimitBytesPerDeployment && storageLimitBytesPerDeployment !== "-1"}
      <div class="quota-entry">
        <div class="quota-entry-title">Data Size</div>
        <div>
          <Progress
            value={totalOrgSize}
            max={Number(storageLimitBytesPerDeployment)}
          />
          {formatDataSizeQuota(totalOrgSize, storageLimitBytesPerDeployment)}
        </div>
      </div>
    {:else}
      <!-- TODO: once we have the dashboard support link to it -->
    {/if}
  {/if}
</div>

<style lang="postcss">
  .quotas {
    @apply flex flex-row items-center gap-x-20 mt-2;
  }

  .quota-entry {
    @apply flex flex-col min-w-24;
  }

  .quota-entry-title {
    @apply font-semibold text-[10px] uppercase text-gray-500;
  }

  .quota-entry-body {
    @apply text-gray-800 text-xs;
  }
</style>
