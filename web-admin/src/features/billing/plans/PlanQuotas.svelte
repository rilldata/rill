<script lang="ts">
  import {
    createAdminServiceListProjectsForOrganization,
    type V1OrganizationQuotas,
  } from "@rilldata/web-admin/client";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { formatDataSizeQuota } from "@rilldata/web-admin/features/billing/plans/utils";
  import { Progress } from "@rilldata/web-common/components/progress";

  export let organization: string;
  export let organizationQuotas: V1OrganizationQuotas;

  $: projects = createAdminServiceListProjectsForOrganization(organization);

  $: projectQuota =
    organizationQuotas.projects && organizationQuotas.projects !== -1
      ? organizationQuotas.projects
      : "Unlimited";

  $: usageMetrics = getOrganizationUsageMetrics(organization);
  $: total = $usageMetrics?.data?.reduce((s, m) => s + m.size, 0) ?? 0;

  $: singleProjectLimit = organizationQuotas.projects === 1;
</script>

<div class="quotas">
  <div class="quota-entry">
    <div class="quota-entry-title">Projects</div>
    <div class="quota-entry-body">
      {$projects.data?.projects?.length} of {projectQuota}
    </div>
  </div>

  <div class="quota-entry">
    <div class="quota-entry-title">Data Size</div>
    <div>
      {#if singleProjectLimit && organizationQuotas.storageLimitBytesPerDeployment !== "-1"}
        <Progress
          value={total}
          max={Number(organizationQuotas.storageLimitBytesPerDeployment)}
        />
        {formatDataSizeQuota(
          total,
          organizationQuotas.storageLimitBytesPerDeployment,
        )}
      {:else}
        <!-- TODO: once we have the dashboard support link to it -->
      {/if}
    </div>
  </div>
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
