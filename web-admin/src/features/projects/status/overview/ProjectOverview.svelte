<script lang="ts">
  import DeploymentSection from "./DeploymentSection.svelte";
  import ErrorsSection from "./ErrorsSection.svelte";
  import AnalyticsSection from "./AnalyticsSection.svelte";

  export let organization: string;
  export let project: string;
</script>

<div class="flex flex-col gap-6">
  <DeploymentSection {organization} {project} />
  <ErrorsSection />
  <AnalyticsSection />
</div>
