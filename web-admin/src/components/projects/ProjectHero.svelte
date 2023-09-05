<script lang="ts">
  import ProjectAccessControls from "./ProjectAccessControls.svelte";
  import ProjectDeploymentStatus from "./ProjectDeploymentStatus.svelte";
  import ProjectGithubConnection from "./ProjectGithubConnection.svelte";
  import ShareProjectCta from "./ShareProjectCTA.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col gap-y-1">
  <h1 class="text-sky-950 text-4xl font-light leading-10">
    {project}
  </h1>
  <ProjectAccessControls {organization} {project}>
    <svelte:fragment slot="manage-project">
      <div
        class="flex flex-wrap gap-x-16 gap-y-8 items-start mt-8 mb-12 rounded w-fit"
      >
        <ProjectDeploymentStatus {organization} {project} />
        <ProjectGithubConnection {organization} {project} />
        <ShareProjectCta {organization} {project} />
      </div>
    </svelte:fragment>
  </ProjectAccessControls>
</div>
