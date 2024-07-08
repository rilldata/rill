<script lang="ts">
  import { page } from "$app/stores";
  import { createAdminServiceRequestProjectAccess } from "@rilldata/web-admin/client";
  import { Button } from "@rilldata/web-common/components/button";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  $: organization = $page.url.searchParams.get("organization");
  $: project = $page.url.searchParams.get("project");

  $: requestAccess = createAdminServiceRequestProjectAccess();
  function onRequestAccess() {
    requested = true;
    void $requestAccess.mutateAsync({
      organization,
      project,
      data: {},
    });
  }

  let requested = false;
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    {#if !requested}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">
          You do not have access to this project.
        </h2>
        <Button type="primary" on:click={onRequestAccess}>Request Access</Button
        >
      </div>
    {:else if $requestAccess.isLoading}
      <div class="h-36 mt-10">
        <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
      </div>
    {:else if $requestAccess.error}
      <div class="flex flex-col gap-y-2">
        <h2 class="text-lg font-semibold">Unable to request access</h2>
        <CtaMessage>
          {$requestAccess.error}
        </CtaMessage>
      </div>
    {:else}
      <div>Requested access</div>
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
