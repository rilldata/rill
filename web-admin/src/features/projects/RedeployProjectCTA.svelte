<script lang="ts">
  import { goto } from "$app/navigation";
  import { orgHasBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;

  $: orgBlockerIssues = orgHasBlockerIssues(organization);
  $: if ($orgBlockerIssues.data) {
    // if projects were hibernated due to a blocker issue on org then take the user to projects page
    void goto(`/${organization}`);
  }
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <CtaHeader variant="bold">Your project is hibernating</CtaHeader>
        <CtaMessage>
          To redeploy the project, run the following command in the Rill CLI:
        </CtaMessage>
        <CLICommandDisplay
          command="rill project hibernate {project} --redeploy"
        />
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
