<script lang="ts">
  import { getBillingStatsForOrg } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  let { organization }: { organization: string } = $props();

  let billingStats = $derived(getBillingStatsForOrg(organization));

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
<div class="flex items-center gap-4">
  <div class="stat-column">
    <span class="stat-value">{fmtCredit($billingStats.prodDailyCost)}/{m.billing_per_day()}</span>
    <span class="stat-label">{m.billing_prod_compute_units({ count: String($billingStats.prodSlots) })}</span>
  </div>
  <div class="stat-column">
    <span class="stat-value">{fmtCredit($billingStats.devDailyCost)}/{m.billing_per_day()}</span>
    <span class="stat-label">{m.billing_dev_compute_units({ count: String($billingStats.devSlots) })}</span>
  </div>
</div>

<style lang="postcss">
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
