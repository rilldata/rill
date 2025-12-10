<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ProjectConfiguration from "@rilldata/web-admin/features/projects/status/ProjectConfiguration.svelte";
  import ProjectParseErrors from "@rilldata/web-admin/features/projects/status/ProjectParseErrors.svelte";
  import ProjectResources from "@rilldata/web-admin/features/projects/status/ProjectResources.svelte";
</script>

<ContentContainer title="Project status" maxWidth={1100}>
  <div class="flex flex-col gap-y-6 size-full">
    <ProjectConfiguration />
    <ProjectResources />
    <ProjectParseErrors />
  </div>
</ContentContainer>
