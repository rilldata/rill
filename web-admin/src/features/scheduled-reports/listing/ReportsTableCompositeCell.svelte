<script lang="ts">
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import cronstrue from "cronstrue";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import { formatRunDate } from "../tableUtils";
  import ReportOwnerBullet from "./ReportOwnerBullet.svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let lastRun: string | undefined;
  export let timeZone: string;
  export let frequency: string;
  export let ownerId: string;
  export let lastRunErrorMessage: string | undefined;

  const humanReadableFrequency = cronstrue.toString(frequency);
</script>

<a href={`reports/${id}`} class="flex flex-col gap-y-0.5 group px-4 py-[5px]">
  <div class="flex gap-x-2 items-center">
    <ReportIcon size={"14px"} className="text-slate-500" />
    <div
      class="text-gray-700 text-sm font-semibold group-hover:text-primary-600"
    >
      {title}
    </div>
    {#if lastRun}
      {#if lastRunErrorMessage}
        <CancelCircleInverse className="text-red-500" />
      {:else}
        <CheckCircleOutline className="text-primary-500" />
      {/if}
    {/if}
  </div>
  <div class="flex gap-x-1 text-gray-500 text-xs font-normal">
    {#if !lastRun}
      <span>Hasn't run yet</span>
    {:else}
      <span>Last run {formatRunDate(lastRun, timeZone)}</span>
    {/if}
    <span>•</span>
    <span>{humanReadableFrequency}</span>
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span>•</span>
        <ReportOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  </div>
</a>
