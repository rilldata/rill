<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ProjectInfo from "@rilldata/web-admin/features/projects/status/project-info/ProjectInfo.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/status/ProjectParseErrors.svelte";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/ProjectResources.svelte";
</script>

<ContentContainer title="Project status" maxWidth={1100}>
  <div class="flex flex-col gap-y-6 size-full pt-1">
    <ProjectInfo />
    <ProjectResources />
    <ProjectParseErrors />
  </div>
</ContentContainer>
