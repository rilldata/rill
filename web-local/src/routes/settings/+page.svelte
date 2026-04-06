<script lang="ts">
  import { get } from "svelte/store";
  import { parse, parseDocument } from "yaml";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import {
    ResourceKind,
    SingletonProjectParserName,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import ResourcesOverviewSection from "@rilldata/web-common/features/resources/overview/ResourcesOverviewSection.svelte";
  import ErrorsOverviewSection from "@rilldata/web-common/features/resources/overview/ErrorsOverviewSection.svelte";
  import {
    countByKind,
    groupErrorsByKind,
  } from "@rilldata/web-common/features/resources/overview-utils";
  import {
    createRuntimeServiceGetFile,
    createRuntimeServiceGetResource,
    createRuntimeServiceListResources,
    type V1Resource,
  } from "@rilldata/web-common/runtime-client";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { goto } from "$app/navigation";

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

  let displayName = $derived(
    (rillYaml.display_name ?? rillYaml.title ?? "") as string,
  );
  let description = $derived((rillYaml.description ?? "") as string);
  let aiInstructions = $derived((rillYaml.ai_instructions ?? "") as string);

  let editDisplayName = $state("");
  let editDescription = $state("");
  let editAiInstructions = $state("");

  $effect(() => {
    if ($rillYamlQuery.isSuccess) {
      editDisplayName = displayName;
      editDescription = description;
      editAiInstructions = aiInstructions;
    }
  });

  // Add blank lines between top-level YAML blocks (key or comment-then-key)
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
      // Add blank line before a comment block that precedes a top-level key
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
      // Add blank line before a top-level key if previous line isn't blank or a comment
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

  async function updateField(key: string, value: string) {
    const artifact = fileArtifacts.getFileArtifact("/rill.yaml");
    let content = get(artifact.editorContent);
    if (!content) {
      await artifact.fetchContent();
      content = get(artifact.remoteContent);
      if (!content) return;
    }
    const doc = parseDocument(content);
    const isNew = !doc.has(key);
    if (value === "" || value === undefined) {
      doc.delete(key);
    } else {
      doc.set(key, value);
      // Add a comment when ai_instructions is first created
      if (isNew && key === "ai_instructions") {
        const raw = doc.toString();
        const out = raw.replace(
          /^(ai_instructions:)/m,
          "# Custom instructions that guide AI features like the chat assistant and AI-generated dashboards.\n# These instructions are included in every AI prompt for this project.\n$1",
        );
        artifact.updateEditorContent(spacedYaml(out), true);
        await artifact.saveLocalContent();
        return;
      }
    }
    artifact.updateEditorContent(spacedYaml(doc.toString()), true);
    await artifact.saveLocalContent();
  }

  // Resources for overview
  const resourcesQuery = createRuntimeServiceListResources(runtimeClient, {});
  let allResources = $derived(
    ($resourcesQuery.data?.resources ?? []) as V1Resource[],
  );
  let resourceCounts = $derived(countByKind(allResources));

  // Parse errors
  const projectParserQuery = createRuntimeServiceGetResource(
    runtimeClient,
    {
      name: {
        kind: ResourceKind.ProjectParser,
        name: SingletonProjectParserName,
      },
    },
    { query: { refetchOnMount: true, refetchOnWindowFocus: true } },
  );
  let parseErrors = $derived(
    $projectParserQuery.data?.resource?.projectParser?.state?.parseErrors ?? [],
  );

  let erroredResources = $derived(
    allResources.filter(
      (r) =>
        !!r.meta?.reconcileError &&
        r.meta?.name?.kind !== ResourceKind.Component,
    ),
  );
  let errorsByKind = $derived(groupErrorsByKind(erroredResources));
  let totalErrors = $derived(parseErrors.length + erroredResources.length);

  function goToResources(
    statusFilter: string[] = [],
    typeFilter: string[] = [],
  ) {
    const params = new URLSearchParams();
    if (statusFilter.length > 0) params.set("status", statusFilter.join(","));
    if (typeFilter.length > 0) params.set("kind", typeFilter.join(","));
    const search = params.toString();
    void goto(`/settings/resources${search ? `?${search}` : ""}`);
  }
</script>

<svelte:head>
  <title>Rill Developer | Settings</title>
</svelte:head>

<!-- Project Configuration (rill.yaml) -->
<div class="section">
  <div class="section-header">
    <h3 class="section-title">Configuration</h3>
  </div>
  <div class="settings-grid">
    <label class="setting-label" for="display-name">Display Name</label>
    <div class="setting-input">
      <input
        id="display-name"
        type="text"
        bind:value={editDisplayName}
        placeholder="Untitled Rill Project"
        onblur={() => {
          if (editDisplayName !== displayName)
            updateField("display_name", editDisplayName);
        }}
        onkeydown={(e) => {
          if (e.key === "Enter") e.currentTarget.blur();
        }}
      />
    </div>

    <label class="setting-label" for="description">Description</label>
    <div class="setting-input">
      <textarea
        id="description"
        bind:value={editDescription}
        placeholder="Project description"
        rows="2"
        onblur={() => {
          if (editDescription !== description)
            updateField("description", editDescription);
        }}
      />
    </div>

    <label class="setting-label" for="ai-instructions">AI Instructions</label>
    <div class="setting-input">
      <textarea
        id="ai-instructions"
        bind:value={editAiInstructions}
        placeholder="Extra instructions for AI features..."
        rows="3"
        onblur={() => {
          if (editAiInstructions !== aiInstructions)
            updateField("ai_instructions", editAiInstructions);
        }}
      />
    </div>
  </div>
</div>

<!-- Resources Overview -->
<ResourcesOverviewSection
  {resourceCounts}
  onViewAll={() => goToResources()}
  onChipClick={(kind) => goToResources([], [kind])}
/>

<!-- Errors Overview -->
<ErrorsOverviewSection
  parseErrorCount={parseErrors.length}
  {errorsByKind}
  {totalErrors}
  isLoading={$projectParserQuery.isLoading || $resourcesQuery.isLoading}
  isError={$projectParserQuery.isError || $resourcesQuery.isError}
  onSectionClick={() => goToResources(["error"])}
  onParseErrorChipClick={() => goToResources(["error"])}
  onKindChipClick={(kind) => goToResources(["error"], [kind])}
/>

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
  .settings-grid {
    @apply grid gap-x-6 gap-y-4 items-start;
    grid-template-columns: 140px 1fr;
  }
  .setting-label {
    @apply text-sm text-fg-secondary pt-2;
  }
  .setting-input input,
  .setting-input textarea {
    @apply w-full px-3 py-2 text-sm rounded-md border bg-surface-base text-fg-primary;
    @apply focus:outline-none focus:ring-1 focus:ring-primary-500;
  }
  .setting-input textarea {
    @apply resize-y;
  }
</style>
