<script lang="ts">
  import { getOrgBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;

  $: orgBlockerIssues = getOrgBlockerIssues(organization);
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <CtaHeader variant="bold">Your project is hibernating</CtaHeader>
        {#if $orgBlockerIssues.data}
          <p class="text-base text-red-600 text-center">
            {$orgBlockerIssues.data} (TODO)
          </p>
        {:else}
          <CtaMessage>
            To redeploy the project, run the following command in the Rill CLI:
          </CtaMessage>
          <CLICommandDisplay
            command="rill project hibernate {project} --redeploy"
          />
        {/if}
      </svelte:fragment>
      <svelte:fragment slot="read-project">
        <CtaHeader variant="bold">This project is hibernating</CtaHeader>
        <CtaMessage>
          Contact the project's administrator to redeploy the project.
        </CtaMessage>
      </svelte:fragment>
    </ProjectAccessControls>
  </CtaContentContainer>
</CtaLayoutContainer>
