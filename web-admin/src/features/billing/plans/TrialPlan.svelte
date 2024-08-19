<script lang="ts">
  import {
    createAdminServiceUpdateBillingSubscription,
    type V1BillingPlan,
    type V1Subscription,
  } from "@rilldata/web-admin/client";
  import PlanContainer from "@rilldata/web-admin/features/billing/plans/PlanContainer.svelte";
  import { getCategorisedPlans } from "@rilldata/web-admin/features/billing/plans/selectors";
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;
  export let plan: V1BillingPlan;
  export let subscription: V1Subscription;

  $: categorisedPlans = getCategorisedPlans();
  $: teamPlan = $categorisedPlans.data.teamPlan;

  $: planUpdater = createAdminServiceUpdateBillingSubscription();
  async function handleUpgradePlan() {
    if (!teamPlan) return;

    await $planUpdater.mutateAsync({
      orgName: organization,
      data: {
        planName: teamPlan.name,
      },
    });
  }

  let open = false;
</script>

<PlanContainer {organization} {plan}>
  <svelte:fragment slot="description">
    Your trial expires in {subscription.trialEndDate}. Ready to get started with
    Rill?
    <a href="https://www.rilldata.com/pricing">See pricing details -></a>
  </svelte:fragment>
  <svelte:fragment slot="contact">
    <span>For custom enterprise needs,</span>
    <Button type="link" compact forcedStyle="padding-left:2px !important;">
      contact us
    </Button>
  </svelte:fragment>

  {#if teamPlan}
    <AlertDialog bind:open slot="action">
      <AlertDialogTrigger asChild let:builder>
        <Button builders={[builder]} type="primary">
          End trial and start Team plan
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Start Team plan</AlertDialogTitle>

          <AlertDialogDescription>
            Your trial will end and your billing cycle will start today. Pricing
            is based on amount of data ingested (and compressed) into Rill.
            <a href="https://www.rilldata.com/pricing">See pricing details -></a
            >
            <ul>
              <li>
                Starts at $250/month with 10 GB included, $25/GB thereafter
              </li>
              <li>Unlimited projects, limited to 50 GB each</li>
            </ul>
          </AlertDialogDescription>

          {#if $planUpdater.error}
            <div class="text-red-500 text-sm py-px">
              {$planUpdater.error.message}
            </div>
          {/if}
        </AlertDialogHeader>
        <AlertDialogFooter class="mt-3">
          <Button type="secondary" on:click={() => (open = false)}>
            Close
          </Button>
          <Button
            type="primary"
            on:click={handleUpgradePlan}
            loading={$planUpdater.isLoading}
          >
            Continue
          </Button>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  {/if}
</PlanContainer>
