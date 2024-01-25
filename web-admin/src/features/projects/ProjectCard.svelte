<script lang="ts">
  import { goto } from "$app/navigation";
  import Card from "@rilldata/web-common/components/card/Card.svelte";
  import Globe from "@rilldata/web-common/components/icons/Globe.svelte";
  import Lock from "@rilldata/web-common/components/icons/Lock.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { createAdminServiceGetProject } from "../../client";
  import ProjectAccessControls from "./ProjectAccessControls.svelte";

  export let organization: string;
  export let project: string;

  // Check whether project is public or private
  $: proj = createAdminServiceGetProject(organization, project);

  function doesProjectNameIncludeUnderscores(project: string) {
    return project.includes("_");
  }
</script>

{#if $proj.data}
  <Card
    on:click={() => goto(`/${organization}/${project}`)}
    bgClasses="bg-gradient-to-b from-white to-slate-50"
  >
    <!-- Project name -->
    <h2
      class="text-gray-700 font-medium text-lg text-center px-4 {doesProjectNameIncludeUnderscores(
        project,
      )
        ? 'break-all'
        : 'break-words'}"
    >
      {project}
    </h2>
    <!-- Permissions tag -->
    <Tag>
      <ProjectAccessControls {organization} {project}>
        <svelte:fragment slot="read-project">Viewer</svelte:fragment>
        <svelte:fragment slot="manage-project">Admin</svelte:fragment>
      </ProjectAccessControls>
    </Tag>
    <!-- Public vs Private indicator -->
    <div class="absolute bottom-2.5 right-2.5 text-slate-400">
      <Tooltip distance={10}>
        {#if $proj.data.project.public}
          <Globe size="16px" />
        {:else}
          <Lock size="16px" />
        {/if}
        <TooltipContent slot="tooltip-content">
          <span class="text-xs"
            >This project is
            {#if $proj.data.project.public}
              <span class="font-medium"> public</span>
            {:else}
              <span class="font-medium"> private</span>
            {/if}
          </span>
        </TooltipContent>
      </Tooltip>
    </div>
  </Card>
{/if}
