<script lang="ts">
  import type { V1Project } from "@rilldata/web-admin/client";
  import ProjectCard from "./ProjectCard.svelte";

  export let organization: string;
  export let projects: V1Project[];
</script>

<ol class="flex gap-6 flex-wrap">
  {#each projects as proj (proj.name)}
    <li>
      <ProjectCard {organization} project={proj.name ?? ""} />
    </li>
  {/each}
</ol>
