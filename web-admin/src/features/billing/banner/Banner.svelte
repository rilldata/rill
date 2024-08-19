<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListPublicBillingPlans,
  } from "@rilldata/web-admin/client";

  export let organization: string;

  $: subscription = createAdminServiceGetBillingSubscription(organization);
  const plans = createAdminServiceListPublicBillingPlans();
  $: console.log($subscription.data?.subscription, $plans.data);
</script>
