<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { page } from "$app/stores";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import { isEmbedPage } from "@rilldata/web-common/layout/navigation/navigation-utils.ts";

  export let branch: string | undefined = undefined;

  const onEmbedPage = isEmbedPage($page);
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
    <CtaHeader variant="bold">
      {#if branch}
        {m.project_starting_branch_deployment()}
      {:else}
        {m.project_deploying()}
      {/if}
    </CtaHeader>
    {#if !onEmbedPage}
      <CtaNeedHelp />
    {/if}
  </CtaContentContainer>
</CtaLayoutContainer>
