<script lang="ts">
  import CLICommandDisplay from "@rilldata/web-common/components/commands/CLICommandDisplay.svelte";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="text-gray-800 flex flex-col gap-y-1">
  <ProjectAccessControls {organization} {project}>
    <span slot="manage-project">
      This project does not currently have a deployment. To redeploy the
      project, run this command in the Rill CLI:
      <CLICommandDisplay command="rill project reconcile --project {project}" />
    </span>
    <span slot="read-project">
      This project is not currently deployed. Contact your project's
      administrator to redeploy the project.
    </span>
  </ProjectAccessControls>
</div>
