<script lang="ts">
  import { useDashboardsLastUpdated } from "@rilldata/web-admin/features/dashboards/listing/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { createAdminServiceGetProject } from "../../client";
  import ProjectDeploymentStatusChip from "./ProjectDeploymentStatusChip.svelte";

  export let organization: string;
  export let project: string;

  $: proj = createAdminServiceGetProject(organization, project);
  $: isProjectDeployed = $proj?.data && $proj.data.prodDeployment;
  $: lastUpdated = useDashboardsLastUpdated(
    $runtime.instanceId,
    organization,
    project
  );
</script>

<div class="flex flex-col gap-y-2">
  <div>
    <span class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
      >Project status</span
    >
    <div>
      <ProjectDeploymentStatusChip {organization} {project} />
    </div>
    {#if $lastUpdated}
      <span class="text-gray-500 text-[11px] leading-4">
        Synced {$lastUpdated.toLocaleString(undefined, {
          month: "short",
          day: "numeric",
          hour: "numeric",
          minute: "numeric",
        })}
      </span>
    {/if}
  </div>
  {#if !isProjectDeployed}
    <div>This project is not deployed.</div>
  {/if}
</div>
