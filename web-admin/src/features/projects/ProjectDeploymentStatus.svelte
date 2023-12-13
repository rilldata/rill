<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ProjectDeploymentStatusChip from "./ProjectDeploymentStatusChip.svelte";
  import { useProjectDataLastRefreshed } from "./selectors";

  export let organization: string;
  export let project: string;

  $: dataLastRefreshed = useProjectDataLastRefreshed($runtime?.instanceId);
</script>

<div class="flex flex-col gap-y-1">
  <span class="uppercase text-gray-500 font-semibold text-[10px] leading-none"
    >Project status</span
  >
  <div>
    <ProjectDeploymentStatusChip {organization} {project} />
  </div>
  {#if $dataLastRefreshed?.data}
    <span class="text-gray-500 text-[11px] leading-4">
      Data refreshed {$dataLastRefreshed.data.toLocaleString(undefined, {
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
      })}
    </span>
  {/if}
</div>
