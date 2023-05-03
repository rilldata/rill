<script lang="ts">
  import { createAdminServiceGetProject } from "../../client";
  import DeploymentStatusChip from "../home/DeploymentStatusChip.svelte";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
</script>

<div class="flex flex-col gap-y-1">
  <span class="uppercase text-gray-500 font-semibold text-[10px] leading-4"
    >Project status</span
  >
  <div>
    <DeploymentStatusChip {organization} {project} />
  </div>
  {#if $proj && $proj.data && $proj.data.prodDeployment}
    <span class="text-gray-500 text-[11px] leading-4">
      Synced {new Date($proj.data.prodDeployment.updatedOn).toLocaleString(
        undefined,
        {
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
        }
      )}
    </span>
  {/if}
</div>
