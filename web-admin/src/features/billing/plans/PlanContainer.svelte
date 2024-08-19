<script lang="ts">
  import type { V1BillingPlan } from "@rilldata/web-admin/client";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";

  export let organization: string;
  export let plan: V1BillingPlan;
</script>

<div class="plan-container">
  <div class="plan-title">{plan.displayName}</div>
  <div><slot name="description" /></div>
  <PlanQuotas {organization} quotas={plan.quotas} />
  <div class="plan-footer">
    <slot name="contact" />
    {#if $$slots["action"]}
      <div class="grow"></div>
    {/if}
    <slot name="action" />
  </div>
</div>

<style lang="postcss">
  .plan-container {
    @apply w-[800px] border border-slate-200 m-5;
  }

  .plan-title {
    @apply text-lg font-semibold;
  }

  .plan-footer {
    @apply flex flex-row items-center;
    @apply bg-slate-50 text-slate-500 border-t border-slate-200;
  }
</style>
