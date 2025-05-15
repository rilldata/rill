<script lang="ts">
  import CTAContentContainer from "@rilldata/web-common/components/calls-to-action/CTAContentContainer.svelte";
  import CTALayoutContainer from "@rilldata/web-common/components/calls-to-action/CTALayoutContainer.svelte";
  import ProjectSelector from "@rilldata/web-common/features/project/ProjectSelector.svelte";
</script>

<CTALayoutContainer>
  <CTAContentContainer>
    <ProjectSelector
      projects={[
        {
          name: "Project A",
          orgName: "Org A",
          id: "project-a",
        },
        {
          name: "Project B",
          orgName: "Org B",
          id: "project-b",
        },
      ]}
    />
  </CTAContentContainer>
</CTALayoutContainer>
