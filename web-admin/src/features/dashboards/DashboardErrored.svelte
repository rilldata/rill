<script lang="ts">
  import CtaButton from "@rilldata/web-common/components/calls-to-action/CTAButton.svelte";
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import ProjectAccessControls from "../projects/ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col justify-center items-center h-3/5 space-y-6 m-auto">
  <CancelCircleInverse size="7em" className="text-gray-200" />
  <div class="flex flex-col items-center space-y-2">
    <h1 class="text-lg font-semibold">
      Sorry, your dashboard isn't working right now!
    </h1>
    <p class="text-gray-500 text-base">
      <ProjectAccessControls {organization} {project}>
        <svelte:fragment slot="manage-project">
          View project status for errors that may help you find a fix.
        </svelte:fragment>
        <svelte:fragment slot="read-project">
          Contact your project's admin for help.
        </svelte:fragment>
      </ProjectAccessControls>
    </p>
  </div>
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
  <p class="text-gray-500">
    Need help? Reach out to us on <a href="https://discord.gg/2ubRfjC7Rh"
      >Discord</a
    >
  </p>
</div>
