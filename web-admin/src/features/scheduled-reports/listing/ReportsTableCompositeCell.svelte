<script lang="ts">
  import CancelCircleInverse from "@rilldata/web-common/components/icons/CancelCircleInverse.svelte";
  import CheckCircleOutline from "@rilldata/web-common/components/icons/CheckCircleOutline.svelte";
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import { resourceColorMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping";
  import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
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
  const reportColor = resourceColorMapping[ResourceKind.Report];
</script>

<a
  href={`reports/${id}`}
  class="flex flex-col gap-y-1 group px-4 py-2.5 w-full h-full"
>
  <div class="flex gap-x-2 items-center min-h-[20px]">
    <ReportIcon size={"14px"} color={reportColor} />
    <span
      class="text-gray-700 text-sm font-semibold group-hover:text-primary-600 truncate"
    >
      {title}
    </span>
    {#if lastRun}
      {#if lastRunErrorMessage}
        <CancelCircleInverse className="text-red-500 shrink-0" />
      {:else}
        <CheckCircleOutline className="text-primary-500 shrink-0" />
      {/if}
    {/if}
  </div>
  <div
    class="flex gap-x-1 text-gray-500 text-xs font-normal min-h-[16px] overflow-hidden"
  >
    {#if !lastRun}
      <span class="shrink-0">Hasn't run yet</span>
    {:else}
      <span class="shrink-0">Last run {formatRunDate(lastRun, timeZone)}</span>
    {/if}
    <span class="shrink-0">•</span>
    <span class="shrink-0 truncate">{humanReadableFrequency}</span>
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span class="shrink-0">•</span>
        <ReportOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  </div>
</a>
