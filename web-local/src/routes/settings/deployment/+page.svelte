<svelte:head>
  <title>Rill Developer | Deployment</title>
</svelte:head>

<div class="section">
  <div class="section-header">
    <h3 class="section-title">Slots</h3>
    <span class="coming-soon">Coming soon</span>
  </div>
  <p class="section-description">
    Configure slot allocation for your project when running Rill Developer in
    the cloud. Slots determine the compute resources available for query
    processing.
  </p>
</div>

<div class="section">
  <div class="section-header">
    <h3 class="section-title">Hibernate</h3>
    <span class="coming-soon">Coming soon</span>
  </div>
  <p class="section-description">
    Configure hibernation settings for your cloud deployment. When enabled, your
    project will automatically hibernate after a period of inactivity to save
    resources and resume when accessed.
  </p>
</div>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5 text-left w-full;
  }
  .section-header {
    @apply flex items-center justify-between mb-3;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }
  .section-description {
    @apply text-sm text-fg-secondary;
  }
  .coming-soon {
    @apply text-xs font-medium text-fg-secondary bg-surface-subtle px-2 py-1 rounded-full;
  }
</style>
