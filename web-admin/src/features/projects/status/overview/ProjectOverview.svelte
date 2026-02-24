<script lang="ts">
  import DeploymentSection from "./DeploymentSection.svelte";
  import ResourcesSection from "./ResourcesSection.svelte";
  import TablesSection from "./TablesSection.svelte";
  import ErrorsSection from "./ErrorsSection.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col gap-6">
  <DeploymentSection {organization} {project} />
  <ResourcesSection />
  <TablesSection />
  <ErrorsSection />
</div>
