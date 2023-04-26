<script lang="ts">
  import DeploymentStatusChip from "../home/DeploymentStatusChip.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col gap-y-1">
  <span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
    >Project status</span
  >
  <div>
    <DeploymentStatusChip {organization} {project} />
  </div>
</div>
