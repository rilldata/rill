<script lang="ts">
  import { page } from "$app/stores";
  import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { onMount } from "svelte";

  /**
   * Landing page to redirect user to stripe payment portal.
   * The link to stripe is a single use hence we need this so that users dont have expired link in email for example.
   */

  $: organization = $page.params.organization;

  onMount(async () => {
    // TODO: once we have mocks we need to add an error state
    window.open(
      await fetchPaymentsPortalURL(
        organization,
        `${$page.url.protocol}//${$page.url.host}/${organization}`,
      ),
      "_self",
    );
  });
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
  </CtaContentContainer>
</CtaLayoutContainer>
