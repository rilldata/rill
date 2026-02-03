<script lang="ts">
  import { onMount } from "svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { LineStatus } from "@rilldata/web-common/components/editor/line-status/state";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Trash from "@rilldata/web-common/components/icons/Trash.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import Resizer from "@rilldata/web-common/layout/Resizer.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    PlusIcon,
    PlayIcon,
    AlertCircleIcon,
    ChevronDownIcon,
  } from "lucide-svelte";
  import { parseDocument, stringify } from "yaml";
  import APIResponsePreview from "./APIResponsePreview.svelte";

  export let fileArtifact: FileArtifact;
  export let errors: LineStatus[];
  export let apiName: string;
  export let switchView: () => void;

  type APIType = "metrics_sql" | "sql" | "api" | "glob" | "resource_status";

  interface Arg {
    id: string;
    key: string;
    value: string;
  }

  // Template options for dropdown menus
  // For metrics_sql/sql: use suffix to append to current base query (select ... from X); content is fallback when empty.
  interface TemplateOption {
    label: string;
    content?: string;
    suffix?: string;
    globPattern?: string;
    api?: string;
    includeArgs?: boolean;
  }

  // Build args for api and glob types (saved to YAML)
  let buildArgs: Arg[] = [];

  const metricsSqlTemplates: TemplateOption[] = [
    {
      label: "Basic Query",
      content: "select measure, dimension from metrics_view",
      suffix: "",
    },
    {
      label: "Query with Filter",
      content:
        "select measure, dimension from metrics_view where dimension = '{{ .args.filter }}'",
      suffix: "where dimension = '{{ .args.filter }}'",
    },
    {
      label: "Query with Limit",
      content:
        "select measure, dimension from metrics_view limit {{ .args.limit }}",
      suffix: "limit {{ .args.limit }}",
    },
    {
      label: "Query with Offset",
      content:
        "select measure, dimension from metrics_view limit {{ .args.limit }} offset {{ .args.offset }}",
      suffix: "limit {{ .args.limit }} offset {{ .args.offset }}",
    },
    {
      label: "Query with Sort",
      content:
        "select measure, dimension from metrics_view order by {{ .args.sort }} {{ .args.order }}",
      suffix: "order by {{ .args.sort }} {{ .args.order }}",
    },
    {
      label: "Full Query",
      content:
        "select measure, dimension from metrics_view where dimension = '{{ .args.filter }}' order by {{ .args.sort }} {{ .args.order }} limit {{ .args.limit }} offset {{ .args.offset }}",
      suffix:
        "where dimension = '{{ .args.filter }}' order by {{ .args.sort }} {{ .args.order }} limit {{ .args.limit }} offset {{ .args.offset }}",
    },
  ];

  const sqlTemplates: TemplateOption[] = [
    {
      label: "Basic Query",
      content: "select * from model_name",
      suffix: "",
    },
    {
      label: "Query with Filter",
      content: "select * from model_name where column = '{{ .args.filter }}'",
      suffix: "where column = '{{ .args.filter }}'",
    },
    {
      label: "Query with Limit",
      content: "select * from model_name limit {{ .args.limit }}",
      suffix: "limit {{ .args.limit }}",
    },
    {
      label: "Query with Offset",
      content:
        "select * from model_name limit {{ .args.limit }} offset {{ .args.offset }}",
      suffix: "limit {{ .args.limit }} offset {{ .args.offset }}",
    },
    {
      label: "Query with Sort",
      content:
        "select * from model_name order by {{ .args.sort }} {{ .args.order }}",
      suffix: "order by {{ .args.sort }} {{ .args.order }}",
    },
    {
      label: "Full Query",
      content:
        "select * from model_name where column = '{{ .args.filter }}' order by {{ .args.sort }} {{ .args.order }} limit {{ .args.limit }} offset {{ .args.offset }}",
      suffix:
        "where column = '{{ .args.filter }}' order by {{ .args.sort }} {{ .args.order }} limit {{ .args.limit }} offset {{ .args.offset }}",
    },
  ];

  const apiTemplates: TemplateOption[] = [
    {
      label: "Basic API Call",
      api: "sample_api",
    },
  ];

  const globTemplates: TemplateOption[] = [
    {
      label: "S3 CSV Files",
      globPattern: "s3://bucket/path/*.csv",
    },
    {
      label: "S3 Parquet Files",
      globPattern: "s3://bucket/path/*.parquet",
    },
    {
      label: "GCS Files",
      globPattern: "gs://bucket/path/*.csv",
    },
    {
      label: "Recursive Pattern",
      globPattern: "s3://bucket/**/*.csv",
    },
    {
      label: "Local Files",
      globPattern: "data/*.csv",
    },
  ];

  // Default templates for each API type (first option from each list)
  const defaultTemplates: Record<
    APIType,
    {
      content?: string;
      globPattern?: string;
      api?: string;
      whereError?: boolean;
    }
  > = {
    metrics_sql: metricsSqlTemplates[0],
    sql: sqlTemplates[0],
    api: apiTemplates[0],
    glob: globTemplates[0],
    resource_status: { whereError: true },
  };

  const apiTypes: { value: APIType; label: string; description: string }[] = [
    {
      value: "metrics_sql",
      label: "Metrics SQL",
      description: "Query metrics views using SQL-like syntax",
    },
    {
      value: "sql",
      label: "SQL",
      description: "Query models/sources using raw SQL",
    },
    {
      value: "api",
      label: "API",
      description: "Call another API endpoint",
    },
    {
      value: "glob",
      label: "Glob",
      description: "Match files using glob patterns",
    },
    {
      value: "resource_status",
      label: "Resource Status",
      description: "Get status of resources with errors",
    },
  ];

  // Extract error messages for display
  $: errorMessages = errors.map((e) => e.message).filter(Boolean);

  $: ({ instanceId } = $runtime);
  $: ({ editorContent, updateEditorContent } = fileArtifact);

  // Parse YAML to extract current type and content
  $: parsedDoc = parseDocument($editorContent ?? "");
  $: currentType = detectAPIType(parsedDoc);
  $: currentContent = extractContent(parsedDoc, currentType);

  let args: Arg[] = [];
  let apiResponse: unknown[] | null = null;
  let responseError: string | null = null;
  let isLoading = false;
  let previewHeight = 300;
  let resizerMax = 600;
  let mainAreaEl: HTMLDivElement;

  onMount(() => {
    if (mainAreaEl) {
      const h = mainAreaEl.clientHeight;
      previewHeight = Math.max(100, Math.floor(h / 2));
      resizerMax = Math.max(400, Math.floor(h * 0.85));
    }
  });

  // Glob-specific fields
  let globPattern = "";

  // API-specific fields
  let targetApiName = "";

  // Resource status fields
  let whereError = true;

  // Connector field for sql and glob
  let connector = "";

  $: host = $runtime.host || "http://localhost:9009";
  $: baseUrl = `${host}/v1/instances/${instanceId}/api/${apiName}`;
  $: fullUrl = buildFullUrl(baseUrl, args);
  $: hasErrors = errors.length > 0;

  function detectAPIType(doc: ReturnType<typeof parseDocument>): APIType {
    if (doc.get("metrics_sql")) return "metrics_sql";
    if (doc.get("sql")) return "sql";
    if (doc.get("glob")) return "glob";
    if (doc.get("resource_status")) return "resource_status";
    const resolver = doc.get("resolver");
    if (resolver === "api") return "api";
    return "metrics_sql"; // default
  }

  function extractContent(
    doc: ReturnType<typeof parseDocument>,
    type: APIType,
  ): string {
    // Extract connector for sql and glob types
    if (type === "sql" || type === "glob") {
      connector = String(doc.get("connector") ?? "");
    }

    // Extract buildArgs for api and glob types
    if (type === "api" || type === "glob") {
      const docJson = doc.toJSON() as Record<string, unknown> | null;
      const argsFromYaml = docJson?.args;
      if (
        argsFromYaml &&
        typeof argsFromYaml === "object" &&
        !Array.isArray(argsFromYaml)
      ) {
        buildArgs = Object.entries(argsFromYaml as Record<string, unknown>).map(
          ([key, value]) => ({
            id: crypto.randomUUID(),
            key,
            value: String(value ?? ""),
          }),
        );
      } else {
        buildArgs = [];
      }
    }

    switch (type) {
      case "metrics_sql":
        return String(doc.get("metrics_sql") ?? "");
      case "sql":
        return String(doc.get("sql") ?? "");
      case "api": {
        targetApiName = String(doc.get("api") ?? "");
        return "";
      }
      case "glob": {
        globPattern = String(doc.get("glob") ?? "");
        return "";
      }
      case "resource_status": {
        // Use getIn to get nested value from YAML document
        const whereErrorValue = doc.getIn(["resource_status", "where_error"]);
        // Default to true unless explicitly set to false
        whereError = whereErrorValue !== false;
        return "";
      }
      default:
        return "";
    }
  }

  function updateYAML(type: APIType, content: string) {
    const doc = parseDocument($editorContent ?? "");

    // Clear old type-specific fields
    doc.delete("metrics_sql");
    doc.delete("sql");
    doc.delete("resolver");
    doc.delete("resolver_properties");
    doc.delete("api");
    doc.delete("args");
    doc.delete("glob");
    doc.delete("resource_status");
    doc.delete("connector");
    doc.delete("path");

    // Set type
    doc.set("type", "api");

    // Set new type-specific fields
    switch (type) {
      case "metrics_sql":
        doc.set("metrics_sql", content);
        break;
      case "sql":
        doc.set("sql", content);
        if (connector) {
          doc.set("connector", connector);
        }
        break;
      case "api":
        doc.set("resolver", "api");
        doc.set("api", targetApiName);
        if (buildArgs.length > 0) {
          const argsObj: Record<string, string> = {};
          buildArgs.forEach((arg) => {
            if (arg.key.trim()) {
              argsObj[arg.key] = arg.value;
            }
          });
          if (Object.keys(argsObj).length > 0) {
            doc.set("args", argsObj);
          }
        }
        break;
      case "glob":
        doc.set("glob", globPattern);
        if (connector) {
          doc.set("connector", connector);
        }
        if (buildArgs.length > 0) {
          const argsObj: Record<string, string> = {};
          buildArgs.forEach((arg) => {
            if (arg.key.trim()) {
              argsObj[arg.key] = arg.value;
            }
          });
          if (Object.keys(argsObj).length > 0) {
            doc.set("args", argsObj);
          }
        }
        break;
      case "resource_status":
        doc.set("resource_status", { where_error: whereError });
        break;
    }

    updateEditorContent(stringify(doc));
  }

  function handleTypeChange(newType: APIType) {
    // When switching types, use default template for the new type
    const defaults = defaultTemplates[newType];

    // Reset connector when switching types
    if (newType === "sql" || newType === "glob") {
      connector = "";
    }

    if (newType === "glob") {
      globPattern = defaults.globPattern ?? "s3://bucket/path/*.csv";
      buildArgs = [];
    } else if (newType === "api") {
      targetApiName = defaults.api ?? "sample_api";
      buildArgs = [];
    } else if (newType === "resource_status") {
      whereError = defaults.whereError ?? true;
    }

    updateYAML(newType, defaults.content ?? "");
  }

  function handleContentChange(content: string) {
    updateYAML(currentType, content);
  }

  function handleGlobPatternChange(value: string) {
    globPattern = value;
    updateYAML("glob", "");
  }

  function handleApiNameChange(value: string) {
    targetApiName = value;
    updateYAML("api", "");
  }

  function handleWhereErrorChange(checked: boolean) {
    whereError = checked;
    updateYAML("resource_status", "");
  }

  function handleConnectorChange(value: string) {
    connector = value;
    updateYAML(currentType, currentContent);
  }

  function addBuildArg() {
    buildArgs = [...buildArgs, { id: crypto.randomUUID(), key: "", value: "" }];
  }

  function removeBuildArg(id: string) {
    buildArgs = buildArgs.filter((arg) => arg.id !== id);
    updateYAML(currentType, currentContent);
  }

  function handleBuildArgChange() {
    updateYAML(currentType, currentContent);
  }

  /** Returns the base query (select ... from X) without where/order by/limit/offset clauses. */
  function getBaseQuery(content: string): string {
    const trimmed = content.trim();
    if (!trimmed) return "";
    const lower = trimmed.toLowerCase();
    const indices = [
      lower.indexOf(" where "),
      lower.indexOf(" order by "),
      lower.indexOf(" limit "),
      lower.indexOf(" offset "),
    ]
      .filter((i) => i >= 0)
      .sort((a, b) => a - b);
    const end = indices.length > 0 ? indices[0] : trimmed.length;
    return trimmed.slice(0, end).trim();
  }

  const defaultBaseQuery: Record<"metrics_sql" | "sql", string> = {
    metrics_sql: "select measure, dimension from metrics_view",
    sql: "select * from model_name",
  };

  function applyTemplate(template: TemplateOption) {
    if (
      (currentType === "metrics_sql" || currentType === "sql") &&
      (template.suffix !== undefined || template.content !== undefined)
    ) {
      // Keep base SQL (select ... from X); remove any existing clauses and replace with template only
      const base =
        getBaseQuery(currentContent) || defaultBaseQuery[currentType];
      const suffix = template.suffix ?? "";
      const newContent = suffix ? `${base} ${suffix}`.trim() : base;
      updateYAML(currentType, newContent);
    } else if (template.content !== undefined) {
      updateYAML(currentType, template.content);
    } else if (template.globPattern !== undefined) {
      globPattern = template.globPattern;
      updateYAML("glob", "");
    } else if (template.api !== undefined) {
      targetApiName = template.api;
      buildArgs = [];
      updateYAML("api", "");
    }
  }

  function buildFullUrl(base: string, params: Arg[]): string {
    const url = new URL(base);
    params.forEach((arg) => {
      if (arg.key.trim()) {
        url.searchParams.set(arg.key, arg.value);
      }
    });
    return url.toString();
  }

  function addArg() {
    args = [...args, { id: crypto.randomUUID(), key: "", value: "" }];
  }

  function removeArg(id: string) {
    args = args.filter((arg) => arg.id !== id);
  }

  async function testAPI() {
    isLoading = true;
    responseError = null;
    apiResponse = null;

    try {
      const response = await fetch(fullUrl);

      if (!response.ok) {
        const errorText = await response.text();
        try {
          const errorJson = JSON.parse(errorText);
          responseError = errorJson.message || errorJson.error || errorText;
        } catch {
          responseError = errorText;
        }
        return;
      }

      const data = await response.json();
      apiResponse = Array.isArray(data) ? data : [data];
    } catch (e) {
      responseError = e instanceof Error ? e.message : "Unknown error occurred";
    } finally {
      isLoading = false;
    }
  }
</script>

<div class="wrapper">
  <div class="main-area">
    <div class="flex flex-col gap-y-4 flex-1 overflow-auto">
      <!-- BUILD SECTION -->
      <div class="section-group">
        <div class="section-header">
          <span class="section-title">Build</span>
        </div>

        <!-- API Type Selector -->
        <div class="section">
          <InputLabel label="API Type" />
          <div class="grid grid-cols-5 gap-2">
            {#each apiTypes as apiType}
              <button
                class="type-button"
                class:selected={currentType === apiType.value}
                on:click={() => handleTypeChange(apiType.value)}
              >
                <span class="type-label">
                  {apiType.label}
                </span>
                <span class="type-description">
                  {apiType.description}
                </span>
              </button>
            {/each}
          </div>
        </div>

        <!-- Mini Editor based on type -->
        <div class="section">
          {#if currentType === "metrics_sql"}
            <div class="flex items-center justify-between">
              <InputLabel label="Metrics SQL Query" />
              <DropdownMenu.Root>
                <DropdownMenu.Trigger asChild let:builder>
                  <Button type="text" compact small builders={[builder]}>
                    Templates
                    <ChevronDownIcon size="12px" />
                  </Button>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="end">
                  {#each metricsSqlTemplates as template}
                    <DropdownMenu.Item on:click={() => applyTemplate(template)}>
                      {template.label}
                    </DropdownMenu.Item>
                  {/each}
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            </div>
            <textarea
              class="query-editor"
              placeholder="select measure, dimension from metrics_view"
              value={currentContent}
              on:input={(e) => handleContentChange(e.currentTarget.value)}
            />
            <p class="hint">
              Replace <code>metrics_view</code> with an actual metrics view name
              from your project. Use measure and dimension names from that metrics
              view.
            </p>
          {:else if currentType === "sql"}
            <div class="mb-3">
              <InputLabel id="sql-connector" label="Connector (optional)" />
              <Input
                value={connector}
                placeholder="duckdb"
                size="md"
                full
                onInput={(value) => handleConnectorChange(value)}
              />
            </div>
            <div class="flex items-center justify-between">
              <InputLabel id="sql-query" label="SQL Query" />
              <DropdownMenu.Root>
                <DropdownMenu.Trigger asChild let:builder>
                  <Button type="text" compact small builders={[builder]}>
                    Templates
                    <ChevronDownIcon size="12px" />
                  </Button>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="end">
                  {#each sqlTemplates as template}
                    <DropdownMenu.Item on:click={() => applyTemplate(template)}>
                      {template.label}
                    </DropdownMenu.Item>
                  {/each}
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            </div>
            <textarea
              class="query-editor"
              placeholder="select * from model_name"
              value={currentContent}
              on:input={(e) => handleContentChange(e.currentTarget.value)}
            />
            <p class="hint">
              Replace <code>table</code> with an actual model or source name from
              your project.
            </p>
          {:else if currentType === "api"}
            <div class="flex items-center justify-between">
              <InputLabel label="Target API Name" />
              <DropdownMenu.Root>
                <DropdownMenu.Trigger asChild let:builder>
                  <Button type="text" compact small builders={[builder]}>
                    Templates
                    <ChevronDownIcon size="12px" />
                  </Button>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="end">
                  {#each apiTemplates as template}
                    <DropdownMenu.Item on:click={() => applyTemplate(template)}>
                      {template.label}
                    </DropdownMenu.Item>
                  {/each}
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            </div>
            <Input
              value={targetApiName}
              placeholder="other_api_name"
              size="lg"
              full
              onInput={(value) => handleApiNameChange(value)}
            />
            <p class="hint">
              Call another API endpoint by name. The target API must exist in
              your project.
            </p>
            <!-- Args for API -->
            <div class="mt-4">
              <div class="flex items-center justify-between">
                <InputLabel id="api-args" label="Args (optional)" />
                <Button type="text" compact small onClick={addBuildArg}>
                  <PlusIcon size="14px" />
                  Add
                </Button>
              </div>
              {#if buildArgs.length === 0}
                <p class="hint">No args added.</p>
              {:else}
                <div class="flex flex-col gap-y-2 mt-2">
                  {#each buildArgs as arg (arg.id)}
                    <div class="flex items-center gap-x-2">
                      <Input
                        bind:value={arg.key}
                        placeholder="Key"
                        size="md"
                        width="180px"
                        onBlur={() => handleBuildArgChange()}
                      />
                      <Input
                        bind:value={arg.value}
                        placeholder="Value"
                        size="md"
                        full
                        onBlur={() => handleBuildArgChange()}
                      />
                      <Button
                        type="ghost"
                        square
                        compact
                        onClick={() => removeBuildArg(arg.id)}
                      >
                        <Trash size="14px" />
                      </Button>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>
          {:else if currentType === "glob"}
            <div class="mb-3">
              <InputLabel id="glob-connector" label="Connector (optional)" />
              <Input
                value={connector}
                placeholder="s3"
                size="md"
                full
                onInput={(value) => handleConnectorChange(value)}
              />
            </div>
            <div class="flex items-center justify-between">
              <InputLabel id="glob-pattern" label="Glob Pattern" />
              <DropdownMenu.Root>
                <DropdownMenu.Trigger asChild let:builder>
                  <Button type="text" compact small builders={[builder]}>
                    Templates
                    <ChevronDownIcon size="12px" />
                  </Button>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="end">
                  {#each globTemplates as template}
                    <DropdownMenu.Item on:click={() => applyTemplate(template)}>
                      {template.label}
                    </DropdownMenu.Item>
                  {/each}
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            </div>
            <Input
              value={globPattern}
              placeholder="s3://bucket/path/*.csv"
              size="lg"
              full
              onInput={(value) => handleGlobPatternChange(value)}
            />
            <p class="hint">
              Match files using glob patterns. Examples: s3://bucket/path/*.csv,
              gs://bucket/*.parquet, s3://bucket/**/*.csv
            </p>
          {:else if currentType === "resource_status"}
            <InputLabel label="Resource Status Options" />
            <div class="flex items-center gap-x-2 mt-1">
              <Checkbox
                checked={whereError}
                onCheckedChange={(checked) => {
                  if (typeof checked === "boolean") {
                    handleWhereErrorChange(checked);
                  }
                }}
              />
              <span class="text-sm">Only show resources with errors</span>
            </div>
            <p class="hint">
              Get the status of resources in your project. Enable the checkbox
              to filter to only resources with errors.
            </p>
          {/if}
        </div>

        <!-- Error Display -->
        {#if errorMessages.length > 0}
          <div class="error-banner">
            <AlertCircleIcon size="16px" class="flex-shrink-0" />
            <div class="flex flex-col gap-y-1">
              {#each errorMessages as message}
                <span>{message}</span>
              {/each}
            </div>
          </div>
        {/if}
      </div>

      <!-- TEST SECTION -->
      <div class="section-group test-section">
        <div class="section-header">
          <span class="section-title">Test</span>
        </div>

        <!-- Endpoint URL -->
        <div class="section">
          <InputLabel label="Endpoint URL" />
          <div
            class="flex items-center gap-x-2 px-3 py-2 bg-surface-muted rounded-[2px] border text-sm font-mono"
          >
            <span class="text-fg-muted">GET</span>
            <span class="flex-1 truncate text-fg-primary">{fullUrl}</span>
          </div>
        </div>

        <!-- Arguments -->
        <div class="section">
          <div class="flex items-center justify-between">
            <InputLabel label="Arguments" />
            <Button type="text" compact onClick={addArg}>
              <PlusIcon size="14px" />
              Add
            </Button>
          </div>

          {#if args.length === 0}
            <p class="hint">
              No arguments added. Click "Add" to add query parameters.
            </p>
          {:else}
            <div class="flex flex-col gap-y-2">
              {#each args as arg (arg.id)}
                <div class="flex items-center gap-x-2">
                  <Input
                    bind:value={arg.key}
                    placeholder="Key"
                    size="md"
                    width="180px"
                  />
                  <Input
                    bind:value={arg.value}
                    placeholder="Value"
                    size="md"
                    full
                  />
                  <Button
                    type="ghost"
                    square
                    compact
                    onClick={() => removeArg(arg.id)}
                  >
                    <Trash size="14px" />
                  </Button>
                </div>
              {/each}
            </div>
          {/if}
        </div>

        <!-- Test Button -->
        <div class="flex items-center gap-x-2">
          <Button
            type="primary"
            onClick={testAPI}
            disabled={hasErrors}
            loading={isLoading}
            loadingCopy="Testing..."
          >
            <PlayIcon size="14px" />
            Test API
          </Button>
          {#if hasErrors}
            <span class="text-sm text-red-500">
              Fix errors above before testing
            </span>
          {/if}
        </div>
      </div>
    </div>

    <!-- Response Preview -->
    <div
      class="preview-area"
      style:height="{previewHeight}px"
      style:min-height="100px"
      style:max-height="60%"
    >
      <Resizer
        absolute={false}
        max={600}
        direction="NS"
        side="top"
        bind:dimension={previewHeight}
      />
      <div class="preview-header">
        <span class="text-xs font-medium text-fg-secondary uppercase">
          Response Preview
        </span>
      </div>
      <div class="preview-content">
        <APIResponsePreview
          response={apiResponse}
          error={responseError}
          {isLoading}
          {apiName}
        />
      </div>
    </div>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    @apply size-full max-w-full max-h-full flex-none;
    @apply overflow-hidden;
    @apply flex;
  }

  .main-area {
    @apply flex flex-col size-full p-4 bg-surface-background border;
    @apply overflow-hidden rounded-[2px];
  }

  .section-group {
    @apply flex flex-col gap-y-4 p-4 rounded-[2px] border bg-surface-background;
  }

  .section-group.test-section {
    @apply bg-surface-background border-dashed;
  }

  .section-header {
    @apply flex items-center gap-x-2 pb-2 border-b mb-2;
  }

  .section-title {
    @apply text-xs font-semibold text-fg-secondary uppercase tracking-wide;
  }

  .section {
    @apply flex flex-col gap-y-2 w-full;
  }

  .type-button {
    @apply flex flex-col items-start p-3 rounded-[2px] border text-left transition-colors;
    @apply bg-surface-subtle;
  }

  .type-button:hover {
    @apply bg-surface-hover;
  }

  .type-button.selected {
    @apply bg-primary-50 border-primary-500;
  }

  .type-label {
    @apply text-sm font-medium;
  }

  .type-button.selected .type-label {
    @apply text-primary-700;
  }

  .type-description {
    @apply text-xs text-fg-muted mt-0.5;
  }

  .query-editor {
    @apply w-full h-56 p-3 font-mono text-sm border rounded-[2px] bg-input resize-none;
    @apply focus:ring-2 focus:ring-primary-100 focus:border-primary-500 outline-none;
  }

  .hint {
    @apply text-xs text-fg-muted;
  }

  .error-banner {
    @apply flex items-start gap-x-2 p-3 rounded-[2px] border mt-4;
    @apply bg-red-50 border-red-200 text-red-700 text-sm;
  }

  .preview-area {
    @apply relative flex flex-col border-t mt-4 flex-shrink-0;
  }

  .preview-header {
    @apply px-3 py-2 border-b bg-surface-subtle;
  }

  .preview-content {
    @apply flex-1 overflow-auto;
  }
</style>
