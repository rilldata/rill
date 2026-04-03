<script lang="ts">
  import ReportIcon from "@rilldata/web-common/components/icons/ReportIcon.svelte";
  import ResourceListRow from "@rilldata/web-common/features/resources/ResourceListRow.svelte";
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

<ResourceListRow
  href={`reports/${id}`}
  {title}
  icon={ReportIcon}
  errorMessage={lastRunErrorMessage}
>
  <svelte:fragment slot="subtitle">
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
  </svelte:fragment>
</ResourceListRow>
