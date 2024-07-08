<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceRejectProjectAccess } from "@rilldata/web-admin/client";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: id = $page.params.id;

  let requested = false;
  $: rejectAccess = createAdminServiceRejectProjectAccess();
  $: if (organization && project && id && !requested) {
    requested = true;
    void $rejectAccess.mutateAsync({
      organization,
      project,
      id,
      data: {},
    });
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $rejectAccess.isLoading}
      <div class="h-36 mt-10">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
    {:else if $rejectAccess.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Unable to reject access</h2>
        <CtaMessage>
          {$rejectAccess.error}
        </CtaMessage>
      </div>
    {:else}
      <div>Access rejected</div>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
