<script lang="ts">
  import { get } from "svelte/store";
  import { parse, parseDocument } from "yaml";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();

  // rill.yaml query
  const rillYamlQuery = createRuntimeServiceGetFile(
    runtimeClient,
    { path: "/rill.yaml" },
    { query: { refetchOnWindowFocus: true } },
  );

  let rillYaml = $derived.by(() => {
    try {
      return (
        parse(($rillYamlQuery.data?.blob as string) ?? "", {
          logLevel: "silent",
        }) ?? {}
      );
    } catch {
      return {};
    }
  });

  let savedFeatures = $derived(
    (rillYaml.features ?? {}) as Record<string, boolean>,
  );
  let hasFeaturesKey = $derived("features" in rillYaml);

  // Feature flag definitions with snake_case keys matching rill.yaml/runtime
  const featureFlagDefs = [
    {
      key: "exports",
      label: "Exports",
      description: "Allow data exports from dashboards",
      default: true,
    },
    {
      key: "alerts",
      label: "Alerts",
      description: "Enable alert creation",
      default: true,
    },
    {
      key: "reports",
      label: "Reports",
      description: "Enable report creation",
      default: true,
    },
    {
      key: "chat",
      label: "Chat",
      description: "Project-level AI chat",
      default: true,
    },
    {
      key: "dashboard_chat",
      label: "Dashboard Chat",
      description: "AI chat within dashboards",
      default: true,
    },
    {
      key: "developer_chat",
      label: "Developer Chat",
      description: "AI chat in the developer editor",
      default: true,
    },
    {
      key: "deploy",
      label: "Deploy",
      description: "Show deploy-related actions",
      default: true,
    },
    {
      key: "dimension_search",
      label: "Dimension Search",
      description: "Global dimension search in dashboards",
      default: false,
    },
    {
      key: "cloud_data_viewer",
      label: "Cloud Data Viewer",
      description: "Source data viewer table in Rill Cloud",
      default: false,
    },
    {
      key: "rill_time",
      label: "Rill Time",
      description: "RillTime syntax range picker",
      default: true,
    },
    {
      key: "sticky_dashboard_state",
      label: "Sticky Dashboard State",
      description: "Persist dashboard state when switching dashboards",
      default: false,
    },
    {
      key: "hide_public_url",
      label: "Hide Public URLs",
      description: "Hide public URL sharing option in dashboards",
      default: false,
    },
    {
      key: "export_header",
      label: "Export Header",
      description: "Include header row in exports",
      default: false,
    },
    {
      key: "two_tiered_navigation",
      label: "Two-Tiered Navigation",
      description: "Use two-tiered navigation layout",
      default: false,
    },
  ];

  // Local draft state
  let featuresEnabled = $state(false);
  let draftFeatures: Record<string, boolean> = $state({});

  // Sync draft from rill.yaml when it changes
  $effect(() => {
    if ($rillYamlQuery.isSuccess) {
      featuresEnabled = hasFeaturesKey;
      draftFeatures = hasFeaturesKey
        ? { ...allDefaults(), ...savedFeatures }
        : {};
    }
  });

  let hasChanges = $derived.by(() => {
    if (featuresEnabled !== hasFeaturesKey) return true;
    if (!featuresEnabled) return false;
    const savedKeys = Object.keys(savedFeatures);
    const draftKeys = Object.keys(draftFeatures);
    if (savedKeys.length !== draftKeys.length) return true;
    return draftKeys.some((k) => savedFeatures[k] !== draftFeatures[k]);
  });

  function getDraftValue(key: string): boolean {
    return draftFeatures[key] ?? false;
  }

  function allDefaults(): Record<string, boolean> {
    const result: Record<string, boolean> = {};
    for (const def of featureFlagDefs) {
      result[def.key] = def.default;
    }
    return result;
  }

  function toggleFeaturesEnabled(enabled: boolean) {
    featuresEnabled = enabled;
    if (enabled) {
      draftFeatures = allDefaults();
    }
  }

  function toggleFeature(key: string) {
    draftFeatures = { ...draftFeatures, [key]: !draftFeatures[key] };
  }

  function spacedYaml(yaml: string): string {
    const lines = yaml.split("\n");
    const result: string[] = [];
    for (let i = 0; i < lines.length; i++) {
      const line = lines[i];
      const isTopLevelKey =
        line.length > 0 &&
        line[0] !== " " &&
        line[0] !== "#" &&
        line[0] !== "-";
      const isComment = line.startsWith("#");
      if (
        isComment &&
        result.length > 0 &&
        result[result.length - 1] !== "" &&
        !result[result.length - 1].startsWith("#")
      ) {
        let j = i;
        while (j < lines.length && lines[j].startsWith("#")) j++;
        if (
          j < lines.length &&
          lines[j].length > 0 &&
          lines[j][0] !== " " &&
          lines[j][0] !== "-"
        ) {
          result.push("");
        }
      }
      if (
        isTopLevelKey &&
        result.length > 0 &&
        result[result.length - 1] !== "" &&
        !result[result.length - 1].startsWith("#")
      ) {
        result.push("");
      }
      result.push(line);
    }
    return result.join("\n");
  }

  async function save() {
    const artifact = fileArtifacts.getFileArtifact("/rill.yaml");
    let content = get(artifact.editorContent);
    if (!content) {
      await artifact.fetchContent();
      content = get(artifact.remoteContent);
      if (!content) return;
    }
    const doc = parseDocument(content);

    const isNew = !doc.has("features");
    if (!featuresEnabled) {
      doc.delete("features");
    } else {
      doc.set("features", doc.createNode(draftFeatures));
    }

    let out = spacedYaml(doc.toString());
    if (isNew && featuresEnabled) {
      out = out.replace(
        /^(features:)/m,
        "# Feature flags to enable or disable specific functionality.\n# Learn more: https://docs.rilldata.com/developers/build/project-configuration#feature-flags\n$1",
      );
    }

    artifact.updateEditorContent(out, true);
    await artifact.saveLocalContent();
  }
</script>

<svelte:head>
  <title>Rill Developer | Developer Settings</title>
</svelte:head>

<div class="section">
  <div class="section-header">
    <h3 class="section-title">Feature Flags</h3>
  </div>

  <div class="enable-row">
    <div class="enable-info">
      <span class="enable-label">Enable feature flags</span>
      <span class="enable-description">
        Override default feature flags for this project. Changes are written to
        <code>rill.yaml</code> under the <code>features:</code> key.
      </span>
    </div>
    <Switch
      small
      checked={featuresEnabled}
      onCheckedChange={(v) => toggleFeaturesEnabled(v)}
    />
  </div>

  {#if featuresEnabled}
    <div class="disclaimer">
      Modifying feature flags can change the behavior of dashboards and AI
      features. Flags not explicitly set will use their default values.
    </div>

    <div class="flags-list">
      {#each featureFlagDefs as { key, label, description } (key)}
        {@const value = getDraftValue(key)}
        <div class="flag-row">
          <div class="flag-info">
            <span class="flag-label">
              {label}
            </span>
            <span class="flag-description">{description}</span>
          </div>
          <Switch
            small
            checked={value}
            onCheckedChange={() => toggleFeature(key)}
          />
        </div>
      {/each}
    </div>
  {/if}

  {#if hasChanges}
    <div class="save-bar">
      <button class="save-button" onclick={save}>Save</button>
    </div>
  {/if}
</div>

<style lang="postcss">
  .section {
    @apply border border-border rounded-lg p-5 text-left w-full;
  }
  .section-header {
    @apply flex items-center justify-between mb-4;
  }
  .section-title {
    @apply text-sm font-semibold text-fg-primary uppercase tracking-wide;
  }

  .enable-row {
    @apply flex items-start justify-between gap-4 py-3;
  }
  .enable-info {
    @apply flex flex-col gap-1;
  }
  .enable-label {
    @apply text-sm font-medium text-fg-primary;
  }
  .enable-description {
    @apply text-xs text-fg-secondary;
  }

  .disclaimer {
    @apply text-xs text-fg-secondary bg-surface-subtle rounded-md px-3 py-2 my-3 border border-border;
  }

  .flags-list {
    @apply flex flex-col divide-y;
  }
  .flag-row {
    @apply flex items-center justify-between gap-4 py-3;
  }
  .flag-info {
    @apply flex flex-col gap-0.5;
  }
  .flag-label {
    @apply text-sm text-fg-primary;
  }
  .flag-description {
    @apply text-xs text-fg-secondary;
  }

  .save-bar {
    @apply flex justify-end pt-4 mt-2 border-t border-border;
  }
  .save-button {
    @apply px-4 py-2 text-sm font-medium rounded-md;
    @apply bg-primary-500 text-white;
  }

  .save-button:hover {
    @apply bg-primary-600 transition-colors;
  }

  code {
    @apply font-mono text-xs bg-surface-subtle px-1 py-0.5 rounded;
  }
</style>
