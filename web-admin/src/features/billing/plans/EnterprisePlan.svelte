<script lang="ts">
  import type { V1OrganizationQuotas } from "@rilldata/web-admin/client";
  import ContactUs from "@rilldata/web-admin/features/billing/ContactUs.svelte";
  import PlanQuotas from "@rilldata/web-admin/features/billing/plans/PlanQuotas.svelte";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";

  export let organization: string;
  export let organizationQuotas: V1OrganizationQuotas;
</script>

<SettingsContainer title="Enterprise plan">
  <div slot="body">
    <div>You’re currently on a custom contract.</div>
    <PlanQuotas {organization} {organizationQuotas} />
  </div>
  <svelte:fragment slot="contact">
    <span>To make changes to your contract,</span>
    <ContactUs variant="enterprise" />
  </svelte:fragment>
</SettingsContainer>
