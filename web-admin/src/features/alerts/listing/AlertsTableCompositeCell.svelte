<script lang="ts">
  import AlertIcon from "@rilldata/web-common/components/icons/AlertIcon.svelte";
  import ResourceListRow from "@rilldata/web-common/features/resources/ResourceListRow.svelte";
  import { timeAgo } from "@rilldata/web-common/lib/time/relative-time";
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import AlertOwnerBullet from "./AlertOwnerBullet.svelte";

  export let organization: string;
  export let project: string;
  export let id: string;
  export let title: string;
  export let lastTrigger: string | undefined;
  export let ownerId: string;
  export let lastTriggerErrorMessage: string | undefined;
</script>

<ResourceListRow
  href={`alerts/${id}`}
  {title}
  icon={AlertIcon}
  errorMessage={lastTriggerErrorMessage}
>
  <svelte:fragment slot="subtitle">
    {#if !lastTrigger}
      <span class="shrink-0">Hasn't been checked yet</span>
    {:else}
      <span class="shrink-0">Last checked {timeAgo(new Date(lastTrigger))}</span>
    {/if}
    <ProjectAccessControls {organization} {project}>
      <svelte:fragment slot="manage-project">
        <span class="shrink-0">•</span>
        <AlertOwnerBullet {organization} {project} {ownerId} />
      </svelte:fragment>
    </ProjectAccessControls>
  </svelte:fragment>
</ResourceListRow>
