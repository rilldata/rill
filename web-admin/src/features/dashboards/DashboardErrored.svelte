<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;

  $: isEmbedded = EmbedStore.isEmbedded();
</script>

<div class="flex flex-col justify-center items-center h-3/5 space-y-6 m-auto">
  <CancelCircleInverse size="7em" className="text-gray-200" />
  <div class="flex flex-col items-center space-y-2">
    <h1 class="text-lg font-semibold">
      {m.dashboard_errored_title()}
    </h1>
    <p class="text-fg-secondary text-base">
      <ProjectAccessControls {organization} {project}>
        <svelte:fragment slot="manage-project">
          {m.dashboard_errored_view_status()}
        </svelte:fragment>
        <svelte:fragment slot="read-project">
          {m.dashboard_errored_contact_admin()}
        </svelte:fragment>
      </ProjectAccessControls>
    </p>
  </div>
  <ProjectAccessControls {organization} {project}>
    <svelte:fragment slot="manage-project">
      <CtaButton
        variant="secondary"
        href={`/${organization}/${project}/-/status`}
        >{m.dashboard_errored_view_status_button()}
      </CtaButton>
    </svelte:fragment>
    <svelte:fragment slot="read-project">
      <CtaButton variant="secondary" href={`/${organization}/${project}`}
        >{m.dashboard_errored_view_project()}
      </CtaButton>
    </svelte:fragment>
  </ProjectAccessControls>
  {#if !isEmbedded}
    <p class="text-fg-secondary">
      {@html m.dashboard_errored_need_help({
        link: `<a href="https://discord.gg/2ubRfjC7Rh">Discord</a>`,
      })}
    </p>
  {/if}
</div>
