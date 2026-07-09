<script lang="ts">
  import { createAdminServiceGetOrganization } from "@rilldata/web-admin/client";
  import { listProjectsForOrgQueryOptions } from "@rilldata/web-admin/features/projects/list-projects-query-options.ts";
  import { createQuery } from "@tanstack/svelte-query";
  import { getOrganizationUsageMetrics } from "@rilldata/web-admin/features/billing/plans/selectors.ts";
  import { formatUsageVsQuota } from "@rilldata/web-admin/features/billing/plans/utils.ts";
  import { Progress } from "@rilldata/web-common/components/progress";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;

  $: projects = createQuery(listProjectsForOrgQueryOptions(organization));
  $: organizationQuotas = createAdminServiceGetOrganization(
    organization,
    undefined,
    {
      query: {
        select: (data) => data.organization?.quotas,
      },
    },
  );

  $: projectQuota =
    $organizationQuotas.data?.projects !== undefined &&
    $organizationQuotas.data?.projects !== -1
      ? $organizationQuotas.data?.projects
      : m.billing_unlimited();

  $: usageMetrics = getOrganizationUsageMetrics(organization);
  $: totalOrgUsage = $usageMetrics?.data?.reduce((s, m) => s + m.size, 0) ?? 0;

  $: singleProjectLimit = $organizationQuotas.data?.projects === 1;
  $: storageLimitBytesPerDeployment =
    $organizationQuotas.data?.storageLimitBytesPerDeployment ?? "";
</script>

<div class="quotas">
  <div class="quota-entry">
    <div class="quota-entry-title">{m.billing_projects()}</div>
    <div class="quota-entry-body">
      {m.billing_x_of_y({ current: String($projects.data?.projects?.length ?? 0), total: String(projectQuota) })}
    </div>
  </div>

  {#if $usageMetrics?.data}
    {#if singleProjectLimit && storageLimitBytesPerDeployment && storageLimitBytesPerDeployment !== "-1"}
      <div class="quota-entry">
        <div class="quota-entry-title">{m.billing_data_size()}</div>
        <div>
          <Progress
            value={totalOrgUsage}
            max={Number(storageLimitBytesPerDeployment)}
          />
          {formatUsageVsQuota(totalOrgUsage, storageLimitBytesPerDeployment)}
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
    @apply font-semibold text-[10px] uppercase text-fg-secondary;
  }

  .quota-entry-body {
    @apply text-fg-primary text-xs;
  }
</style>
