<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import TabNav from "@rilldata/web-common/components/nav/TabNav.svelte";
  import Rocket from "svelte-radix/Rocket.svelte";
  import ResourcesSection from "../status/ResourcesSection.svelte";
  import ParseErrorsSection from "../status/ParseErrorsSection.svelte";

  let selectedPage = "overview";

  const navItems = [
    { label: "Overview", value: "overview" },
    { label: "Resources", value: "resources" },
    { label: "Logs", value: "logs" },
  ];
</script>

<ContentContainer title="Project Status" maxWidth={1100}>
  <div class="flex pt-6 gap-6 max-w-full overflow-hidden">
    <TabNav items={navItems} bind:selected={selectedPage} />

    <!-- Main Content -->
    <div class="flex flex-col gap-y-6 w-full overflow-hidden">
      {#if selectedPage === "overview"}
        <div class="section">
          <div class="section-header">
            <h3 class="section-title">Deployment</h3>
          </div>
          <p class="text-sm text-fg-muted">
            Deploy your project to Rill Cloud to monitor deployments, view logs,
            and manage your project.
          </p>
          <div class="mt-4">
            <Button type="primary" href="/deploy" compact>
              <Rocket size="14px" />
              Deploy to Rill Cloud
            </Button>
          </div>
        </div>
      {:else if selectedPage === "resources"}
        <ResourcesSection />
        <ParseErrorsSection />
      {:else if selectedPage === "logs"}
        <div class="section">
          <div class="section-header">
            <h3 class="section-title">Logs</h3>
          </div>
          <p class="text-sm text-fg-muted">
            Real-time logs are available after deploying to Rill Cloud.
          </p>
          <div class="mt-4">
            <Button type="primary" href="/deploy" compact>
              <Rocket size="14px" />
              Deploy to Rill Cloud
            </Button>
          </div>
        </div>
      {/if}
    </div>
  </div>
</ContentContainer>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
</style>
