<script lang="ts">
  import ProjectDeploymentStatusChip from "./ProjectDeploymentStatusChip.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col gap-y-1">
  <span class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
    >Deployment</span
  >
  <div>
    <ProjectDeploymentStatusChip {organization} {project} />
  </div>
</div>
