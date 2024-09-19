<script lang="ts">
  import {
    createAdminServiceGetBillingSubscription,
    createAdminServiceListOrganizationBillingIssues,
    createAdminServiceListPublicBillingPlans,
  } from "@rilldata/web-admin/client";

  export let organization: string;

  $: subscription = createAdminServiceGetBillingSubscription(organization);
  const plans = createAdminServiceListPublicBillingPlans();
  $: issues = createAdminServiceListOrganizationBillingIssues(organization);
</script>
