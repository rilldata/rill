<script lang="ts">
  import {
    getBillingStatsForOrg,
    getOrganizationUsageMetrics,
  } from "@rilldata/web-admin/features/billing/plans/selectors";
  import InfoCircle from "@rilldata/web-common/components/icons/InfoCircle.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";

  let { organization }: { organization: string } = $props();

  let billingStats = $derived(getBillingStatsForOrg(organization));

  let usageMetrics = $derived(getOrganizationUsageMetrics(organization));
  let totalStorage = $derived(
    $usageMetrics?.data?.reduce((s, m) => s + m.size, 0) ?? 0,
  );

  // Storage cost estimate from current snapshot. Final billing is the
  // average across the cycle, so the invoice may differ; the UI labels
  // this as an estimate via tooltip.
  const BYTES_PER_GB = 1024 ** 3;
  const STORAGE_FREE_GB = 1;
  const STORAGE_RATE_PER_GB = 1;
  let storageCost = $derived(
    Math.max(0, totalStorage / BYTES_PER_GB - STORAGE_FREE_GB) *
      STORAGE_RATE_PER_GB,
  );

  function fmtCredit(n: number): string {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
    }).format(n);
  }
</script>

<!-- Cost + usage row -->
<!-- TODO: replace prod/dev dollar values with accrued costs once the
     billing usage API exposes them. Current values project from
     current config × list rate, which is misleading vs. actual
     billed amounts. Storage is estimated from current snapshot. -->
<div class="stats-row">
  <div class="flex items-center gap-4">
    <div class="stat-column">
      <span class="stat-value">{fmtCredit($billingStats.prodDailyCost)}</span>
      <span class="stat-label"
        >{$billingStats.prodSlots} Prod Compute Units</span
      >
    </div>
    <div class="stat-column">
      <span class="stat-value">{fmtCredit($billingStats.devDailyCost)}</span>
      <span class="stat-label">{$billingStats.devSlots} Dev Compute Units</span>
    </div>
    <div class="stat-column">
      <div class="flex items-center gap-1">
        <span class="stat-value">{fmtCredit(storageCost)}</span>
        <Tooltip location="top" alignment="middle" distance={8}>
          <span class="text-fg-muted flex">
            <InfoCircle size="14px" />
          </span>
          <TooltipContent maxWidth="260px" slot="tooltip-content">
            Estimated from current storage at $1/GB/month after a 1 GB free
            allowance. Final billing is based on average storage across the
            cycle, so the invoice may differ.
          </TooltipContent>
        </Tooltip>
      </div>
      <span class="stat-label"
        >{totalStorage > 0 ? formatMemorySize(totalStorage) : "0 B"} Storage</span
      >
    </div>
  </div>
</div>

<style lang="postcss">
  .stats-row {
    @apply flex items-center justify-between bg-surface-subtle border-t;
    margin: 16px -24px 0;
    padding: 12px 24px;
  }

  .stat-column {
    @apply flex flex-col gap-1;
  }

  .stat-value {
    @apply font-sans font-medium text-sm;
    line-height: 100%;
  }

  .stat-label {
    @apply font-sans font-medium text-xs text-fg-tertiary;
    line-height: 100%;
  }
</style>
