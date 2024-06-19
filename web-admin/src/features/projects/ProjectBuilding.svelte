<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CtaContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CtaHeader from "@rilldata/web-common/components/calls-to-action/CTAHeader.svelte";
  import CtaLayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import CtaNeedHelp from "@rilldata/web-common/components/calls-to-action/CTANeedHelp.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;
</script>

<CtaLayoutContainer>
  <CtaContentContainer>
    <div class="h-36">
      <Spinner status={EntityStatus.Running} size="7rem" duration={725} />
    </div>
    <CtaHeader variant="bold"
      >Hang tight! We're building your dashboard...</CtaHeader
    >
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <CtaButton
          variant="secondary"
          href={`/${organization}/${project}/-/status`}
          >View project status
        </CtaButton>
      </svelte:fragment>
      <svelte:fragment slot="read-project">
        <CtaButton variant="secondary" href={`/${organization}/${project}`}
          >View project
        </CtaButton>
      </svelte:fragment>
    </ProjectAccessControls>
    <CtaNeedHelp />
  </CtaContentContainer>
</CtaLayoutContainer>
