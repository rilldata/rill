<script lang="ts">
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaMessage from "@rilldata/web-common/components/calls-to-action/CTAMessage.svelte";
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;
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
          command="rill project reconcile --project {project}"
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
