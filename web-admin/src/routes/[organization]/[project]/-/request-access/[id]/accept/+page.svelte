<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceAcceptProjectAccess } from "@rilldata/web-admin/client";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: id = $page.params.id;

  let requested = false;
  $: acceptAccess = createAdminServiceAcceptProjectAccess();
  $: if (organization && project && id && !requested) {
    requested = true;
    void $acceptAccess.mutateAsync({
      organization,
      project,
      id,
      data: {},
    });
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if $acceptAccess.isLoading}
      <div class="h-36 mt-10">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
    {:else if $acceptAccess.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Unable to accept access</h2>
        <CtaMessage>
          {$acceptAccess.error}
        </CtaMessage>
      </div>
    {:else}
      <div>Access granted</div>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
