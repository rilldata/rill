<script lang="ts">
  import { createAdminServiceGetBillingSubscription } from "@rilldata/web-admin/client";
  import { getPlanForOrg } from "@rilldata/web-admin/features/billing/selectors";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;

  $: subscription = createAdminServiceGetBillingSubscription(organization);
  $: plan = getPlanForOrg(organization);

  $: isTrial = !!$subscription?.data?.subscription?.trialEndDate;
  $: hasEnded = !!$subscription?.data?.subscription?.endDate;
  $: isBilled = !!$subscription?.data?.subscription?.currentBillingCycleEndDate;
</script>

<div class="w-[800px] border border-slate-200 m-5">
  {#if $subscription.data?.subscription && $plan}
    <div class="flex flex-col p-3">
      <div class="text-lg font-semibold">{$plan.displayName}</div>
      <div>
        {#if isTrial}
          Your trial expires in {$subscription.data.subscription.trialEndDate}.
          Ready to get started with Rill?
          <a href="https://www.rilldata.com/pricing">See pricing details -></a>
        {:else if hasEnded}
          Your subscription ends on {$subscription.data.subscription.endDate}.
        {:else if isBilled}
          Next billing cycle will start on {$subscription.data.subscription
            .currentBillingCycleEndDate}
          <a href="https://www.rilldata.com/pricing">See pricing details -></a>
        {:else}
          You’re currently on a custom contract.
        {/if}
      </div>
    </div>
    <div
      class="flex flex-row items-center p-3 bg-slate-50 text-slate-500 border-t border-slate-200"
    >
      {#if isTrial || isBilled}
        <span>For custom enterprise needs,</span>
        <Button type="link" compact forcedStyle="padding-left:2px !important;"
          >contact us</Button
        >
      {:else if !hasEnded}
        <span>To make changes to your contract,</span>
        <Button type="link" compact forcedStyle="padding-left:2px !important;"
          >contact support</Button
        >
      {/if}
      <div class="grow"></div>
      {#if isTrial}
        <Button type="primary">End trial and start Team plan</Button>
      {:else if hasEnded}
        <Button type="primary">Renew Team plan</Button>
      {:else if isBilled}
        <Button type="secondary">Cancel</Button>
      {/if}
    </div>
  {/if}
</div>
