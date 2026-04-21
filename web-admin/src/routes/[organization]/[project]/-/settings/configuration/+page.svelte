<script lang="ts">
  import { page } from "$app/stores";
  import SettingsContainer from "@rilldata/web-admin/features/organizations/settings/SettingsContainer.svelte";
  import RadixLarge from "@rilldata/web-common/components/typography/RadixLarge.svelte";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { createRuntimeServiceGetInstance } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();

  $: organization = $page.params.organization;
  $: project = $page.params.project;

  // Non-sensitive fields (projectDisplayName, theme, featureFlags, aiInstructions,
  // frontendUrl) are readable by any viewer with ReadObjects.
  $: baseQuery = createRuntimeServiceGetInstance(runtimeClient, {});

  // Sensitive fields (connectors, variables, annotations) require ReadInstance,
  // granted to users with ManageProject. Errors silently when the caller lacks
  // the permission (typical prod viewer).
  $: sensitiveQuery = createRuntimeServiceGetInstance(
    runtimeClient,
    { sensitive: true },
    { query: { retry: false } },
  );

  $: instance = $baseQuery.data?.instance;
  $: sensitive = $sensitiveQuery.data?.instance;
  $: hasSensitive = !!sensitive;

  $: featureFlagEntries = Object.entries(instance?.featureFlags ?? {}).sort(
    ([a], [b]) => a.localeCompare(b),
  );
  $: projectConnectors = sensitive?.projectConnectors ?? [];
  $: annotationEntries = Object.entries(sensitive?.annotations ?? {}).sort(
    ([a], [b]) => a.localeCompare(b),
  );
  $: olapConnector = projectConnectors.find(
    (c) => c.name === sensitive?.olapConnector,
  );
  $: aiConnector = projectConnectors.find(
    (c) => c.name === sensitive?.aiConnector,
  );
</script>

<div class="flex flex-col gap-6 w-full overflow-hidden">
  <div class="flex flex-col">
    <RadixLarge>Project configuration</RadixLarge>
    <p class="text-sm text-fg-tertiary font-medium">
      Project-level settings from <code>rill.yaml</code> as resolved by the
      runtime.
      <a
        href="https://docs.rilldata.com/reference/project-files/rill-yaml"
        target="_blank"
        class="text-primary-600 hover:text-primary-700 active:text-primary-800"
      >
        Learn more ->
      </a>
    </p>
  </div>

  {#if $baseQuery.isLoading}
    <DelayedSpinner isLoading={$baseQuery.isLoading} size="1rem" />
  {:else if $baseQuery.isError}
    <div
      class="flex items-center justify-center border rounded-sm bg-surface-subtle text-fg-tertiary text-sm py-10"
    >
      Failed to load project configuration.
    </div>
  {:else if instance}
    <SettingsContainer title="General">
      <dl class="config-grid">
        <dt>Display name</dt>
        <dd>{instance.projectDisplayName || "—"}</dd>

        <dt>Theme</dt>
        <dd>{instance.theme || "—"}</dd>

        {#if instance.frontendUrl}
          <dt>Frontend URL</dt>
          <dd class="font-mono break-all">{instance.frontendUrl}</dd>
        {/if}
      </dl>
    </SettingsContainer>

    {#if instance.aiInstructions}
      <SettingsContainer title="AI instructions">
        <pre class="ai-instructions">{instance.aiInstructions}</pre>
      </SettingsContainer>
    {/if}

    <SettingsContainer title="Feature flags">
      {#if featureFlagEntries.length === 0}
        <span class="text-fg-tertiary">No feature flags set.</span>
      {:else}
        <ul class="flag-list">
          {#each featureFlagEntries as [name, enabled] (name)}
            <li>
              <span
                class="flag-dot"
                class:flag-dot-on={enabled}
                aria-hidden="true"
              ></span>
              <span class="font-mono">{name}</span>
              <span class="text-fg-tertiary">{enabled ? "on" : "off"}</span>
            </li>
          {/each}
        </ul>
      {/if}
    </SettingsContainer>

    {#if hasSensitive}
      <SettingsContainer title="Connectors">
        <dl class="config-grid">
          {#if olapConnector}
            <dt>OLAP</dt>
            <dd>
              <span class="font-mono">{olapConnector.name}</span>
              <span class="text-fg-tertiary">({olapConnector.type})</span>
            </dd>
          {/if}
          {#if aiConnector}
            <dt>AI</dt>
            <dd>
              <span class="font-mono">{aiConnector.name}</span>
              <span class="text-fg-tertiary">({aiConnector.type})</span>
            </dd>
          {/if}
        </dl>
        {#if projectConnectors.length > 0}
          <div class="mt-4">
            <div class="text-fg-secondary font-semibold text-xs uppercase mb-2">
              All connectors
            </div>
            <ul class="connector-list">
              {#each projectConnectors as connector (connector.name)}
                <li>
                  <span class="font-mono">{connector.name}</span>
                  <span class="text-fg-tertiary">({connector.type})</span>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      </SettingsContainer>

      {#if annotationEntries.length > 0}
        <SettingsContainer title="Annotations">
          <dl class="config-grid">
            {#each annotationEntries as [key, value] (key)}
              <dt class="font-mono">{key}</dt>
              <dd class="font-mono break-all">{value}</dd>
            {/each}
          </dl>
        </SettingsContainer>
      {/if}
    {/if}

    <SettingsContainer title="Project variables">
      <p>
        Manage project variables on the <a
          href={`/${organization}/${project}/-/settings/environment-variables`}
          class="text-primary-600 hover:text-primary-700 active:text-primary-800"
          >Environment Variables</a
        > tab.
      </p>
    </SettingsContainer>
  {/if}
</div>

<style lang="postcss">
  .config-grid {
    @apply grid gap-x-6 gap-y-2;
    grid-template-columns: max-content 1fr;
  }

  .config-grid dt {
    @apply text-fg-secondary font-semibold;
  }

  .config-grid dd {
    @apply text-fg-primary;
  }

  .ai-instructions {
    @apply bg-surface-subtle rounded-sm p-3 text-xs font-mono whitespace-pre-wrap;
  }

  .flag-list {
    @apply flex flex-col gap-y-1;
  }

  .flag-list li {
    @apply flex items-center gap-x-2 text-sm;
  }

  .flag-dot {
    @apply inline-block w-2 h-2 rounded-full bg-fg-tertiary;
  }

  .flag-dot-on {
    @apply bg-primary-600;
  }

  .connector-list {
    @apply flex flex-col gap-y-1;
  }

  .connector-list li {
    @apply flex items-center gap-x-2 text-sm;
  }
</style>
