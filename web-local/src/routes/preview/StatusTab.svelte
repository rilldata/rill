<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Rocket from "svelte-radix/Rocket.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import ResourcesSection from "../status/ResourcesSection.svelte";
  import ParseErrorsSection from "../status/ParseErrorsSection.svelte";

  type StatusPage = "overview" | "resources" | "logs";

  let selectedPage: StatusPage = "overview";

  const navItems: Array<{ label: string; route: StatusPage }> = [
    { label: "Overview", route: "overview" },
    { label: "Resources", route: "resources" },
    { label: "Logs", route: "logs" },
  ];
</script>

<main
  class="bg-surface-base size-full pt-8 pb-16 lg:pt-12 flex flex-col items-center px-8 sm:px-16 lg:px-32 2xl:px-40 overflow-y-auto"
  style="scrollbar-gutter: stable;"
>
  <section class="w-full flex flex-col gap-y-3" style:max-width="1100px">
    <h1
      class="text-2xl text-fg-primary font-semibold"
      aria-label="Container title"
    >
      Project Status
    </h1>

    <div class="layout">
      <!-- Left Navigation (matches admin LeftNav) -->
      <div class="nav-items" style:min-width="180px">
        {#each navItems as item (item.route)}
          <button
            on:click={() => (selectedPage = item.route)}
            class="nav-item"
            class:selected={selectedPage === item.route}
          >
            <span class="text-fg-primary">{item.label}</span>
          </button>
        {/each}
      </div>

      <!-- Main Content -->
      <div class="flex flex-col gap-y-6 w-full overflow-hidden">
        {#if selectedPage === "overview"}
          <div class="section">
            <div class="section-header">
              <h3 class="section-title">Deployment</h3>
            </div>
            <p class="text-sm text-fg-muted">
              Deploy your project to Rill Cloud to monitor deployments, view logs, and manage your project.
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
  </section>
</main>

<style lang="postcss">
  .layout {
    @apply flex pt-6 gap-6 max-w-full overflow-hidden;
  }

  .nav-items {
    @apply flex flex-col gap-y-2;
  }

  .nav-item {
    @apply p-2 flex gap-x-1 items-center;
    @apply rounded-sm;
    @apply text-sm font-medium;
  }

  .selected {
    @apply bg-surface-active;
  }

  .nav-item:hover {
    @apply bg-surface-hover;
  }

  .section {
    @apply border border-border rounded-lg p-5;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .info-grid {
    @apply flex flex-col;
  }
  .info-row {
    @apply flex items-center py-2;
  }
  .info-label {
    @apply text-sm text-fg-secondary w-32 shrink-0;
  }
  .info-value {
    @apply text-sm text-fg-primary;
  }
</style>
