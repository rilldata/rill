<script lang="ts">
  import ProjectAccessControls from "../../projects/ProjectAccessControls.svelte";
  import { useReportOwnerName } from "../selectors";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  $: ownerName = useReportOwnerName(organization, project, ownerId);
</script>

<ProjectAccessControls {organization} {project}>
  <svelte:fragment slot="manage-project">
    {#if $ownerName.isSuccess}
      <span class="truncate">
        {$ownerName.data ?? "Code-defined"}
      </span>
    {:else}
      <span class="text-fg-tertiary">—</span>
    {/if}
  </svelte:fragment>
</ProjectAccessControls>
