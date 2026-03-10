<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import type { V1Resource } from "@rilldata/web-common/runtime-client";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { useRuntimeClient } from "../../../../runtime-client/v2";
  import {
    runtimeServiceDetectPython,
    runtimeServiceGetFile,
    runtimeServicePutFile,
    runtimeServiceSetupPythonEnvironment,
  } from "../../../../runtime-client/v2/gen/runtime-service";
  import { createSource } from "../createSource";
  import { pythonTemplates } from "./templates";

  export let onClose: () => void = () => {};
  export let onBack: () => void = () => {};

  const runtimeClient = useRuntimeClient();

  // ── Wizard step: "detect" → "setup" → "source" ──
  let wizardStep: "detect" | "setup" | "source" = "detect";

  // ── Detect state ──
  let detecting = true;
  let pythonFound = false;
  let pythonPath = "";
  let pythonVersion = "";
  let detectError = "";

  // ── Setup state ──
  let settingUp = false;
  let setupComplete = false;
  let setupError = "";
  let installedPackages: string[] = [];

  // Package templates (setup step)
  const packageTemplates = [
    {
      id: "ga4",
      label: "Google Analytics (GA4)",
      description: "Pull analytics data from GA4 properties",
      packages: ["google-analytics-data"],
      checked: false,
    },
    {
      id: "stripe",
      label: "Stripe",
      description: "Billing, charges, subscriptions, and more",
      packages: ["stripe"],
      checked: false,
    },
    {
      id: "aws",
      label: "AWS / S3",
      description: "AWS services and S3 access",
      packages: ["boto3"],
      checked: false,
    },
    {
      id: "gcp",
      label: "Google Cloud",
      description: "BigQuery, GCS, and other GCP services",
      packages: ["google-cloud-bigquery", "google-cloud-storage"],
      checked: false,
    },
    {
      id: "http",
      label: "REST APIs",
      description: "HTTP requests for any REST endpoint",
      packages: ["requests"],
      checked: false,
    },
    {
      id: "dbt",
      label: "dbt",
      description: "dbt Core with DuckDB adapter",
      packages: ["dbt-core", "dbt-duckdb"],
      checked: false,
    },
    {
      id: "sql",
      label: "SQL Databases",
      description: "Connect to Postgres, MySQL, SQLite",
      packages: ["sqlalchemy", "psycopg2-binary"],
      checked: false,
    },
  ];

  let customPackageInput = "";
  const BASE_PACKAGES = ["pandas", "pyarrow"];

  $: selectedPackages = buildPackageList(packageTemplates, customPackageInput);

  function buildPackageList(
    templates: typeof packageTemplates,
    custom: string,
  ): string[] {
    const pkgs: string[] = [...BASE_PACKAGES];
    for (const t of templates) {
      if (t.checked) {
        for (const p of t.packages) {
          if (!pkgs.includes(p)) pkgs.push(p);
        }
      }
    }
    if (custom.trim()) {
      for (const raw of custom.split(",")) {
        const p = raw.trim();
        if (p && !pkgs.includes(p)) pkgs.push(p);
      }
    }
    return pkgs;
  }

  // ── Source step state ──
  let sourceMode: "template" | "custom" = "template";
  let selectedTemplateId: string | null = null;
  let codePath = "";
  let modelName = "";
  let creating = false;
  let createError = "";

  // ── Connector secrets ──
  const EXCLUDED_DRIVERS = new Set(["python", "duckdb", "file", "https"]);

  const connectorsQuery = useFilteredResources(
    runtimeClient,
    ResourceKind.Connector,
    (data) =>
      (data.resources ?? []).filter(
        (r: V1Resource) =>
          r.connector?.spec?.driver &&
          !EXCLUDED_DRIVERS.has(r.connector.spec.driver),
      ),
  );

  let selectedSecrets: string[] = [];

  function toggleSecret(name: string) {
    if (selectedSecrets.includes(name)) {
      selectedSecrets = selectedSecrets.filter((s) => s !== name);
    } else {
      selectedSecrets = [...selectedSecrets, name];
    }
  }

  function selectTemplate(id: string) {
    const tmpl = pythonTemplates.find((t) => t.id === id);
    if (!tmpl) return;
    selectedTemplateId = id;
    codePath = tmpl.defaultPath;
    const slug = tmpl.defaultPath.split("/").pop()?.replace(/\.py$/, "") ?? "";
    if (slug) modelName = slug.replace(/[^a-zA-Z0-9_]/g, "_");
    // Pre-select suggested secrets
    for (const s of tmpl.suggestedSecrets) {
      if (!selectedSecrets.includes(s)) {
        selectedSecrets = [...selectedSecrets, s];
      }
    }
  }

  // ── Lifecycle ──
  init();

  async function init() {
    await Promise.all([detectPython(), loadExistingRequirements()]);
  }

  async function detectPython() {
    detecting = true;
    detectError = "";
    try {
      const result = await runtimeServiceDetectPython(runtimeClient, {});
      pythonFound = result.found ?? false;
      pythonPath = result.path ?? "";
      pythonVersion = result.version ?? "";
      if (pythonFound) {
        wizardStep = "setup";
      }
    } catch (err) {
      detectError = err instanceof Error ? err.message : String(err);
    } finally {
      detecting = false;
    }
  }

  async function loadExistingRequirements() {
    try {
      const file = await runtimeServiceGetFile(runtimeClient, {
        path: "requirements.txt",
      });
      const content = file.blob ?? "";
      if (!content.trim()) return;

      const existingPkgs = content
        .split("\n")
        .map((line: string) => line.trim())
        .filter(
          (line: string) =>
            line && !line.startsWith("#") && !line.startsWith("-"),
        );

      for (let i = 0; i < packageTemplates.length; i++) {
        const allPresent = packageTemplates[i].packages.every((p) =>
          existingPkgs.some(
            (ep: string) =>
              ep === p || ep.startsWith(p + "==") || ep.startsWith(p + ">="),
          ),
        );
        if (allPresent) {
          packageTemplates[i].checked = true;
        }
      }

      const templatePkgs = new Set(
        packageTemplates.flatMap((t) => t.packages),
      );
      const customPkgs = existingPkgs.filter(
        (p: string) =>
          !BASE_PACKAGES.includes(p.split("==")[0].split(">=")[0]) &&
          !templatePkgs.has(p.split("==")[0].split(">=")[0]),
      );
      if (customPkgs.length > 0) {
        customPackageInput = customPkgs.join(", ");
      }
    } catch {
      // No requirements.txt yet
    }
  }

  async function setupEnvironment() {
    settingUp = true;
    setupError = "";
    try {
      const extraPackages = selectedPackages.filter(
        (p) => !BASE_PACKAGES.includes(p),
      );
      const result = await runtimeServiceSetupPythonEnvironment(
        runtimeClient,
        {
          packages: extraPackages,
          pythonPath: pythonPath || undefined,
        },
      );
      installedPackages = (result.installedPackages as string[]) ?? [];
      setupComplete = true;
      wizardStep = "source";
    } catch (err) {
      setupError = err instanceof Error ? err.message : String(err);
    } finally {
      settingUp = false;
    }
  }

  async function createModel() {
    if (!modelName.trim()) return;
    if (sourceMode === "template" && !selectedTemplateId) return;
    if (sourceMode === "custom" && !codePath.trim()) return;

    creating = true;
    createError = "";
    try {
      // If template: write the script file into the project
      if (sourceMode === "template" && selectedTemplateId) {
        const tmpl = pythonTemplates.find((t) => t.id === selectedTemplateId);
        if (tmpl) {
          codePath = tmpl.defaultPath;
          await runtimeServicePutFile(runtimeClient, {
            path: tmpl.defaultPath,
            blob: tmpl.script,
            create: true,
          });
        }
      }

      // Build model YAML
      const lines = [
        "# Model YAML",
        "# Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/python",
        "",
        "type: model",
        "materialize: true",
        "",
        "connector: python",
        "",
        `code_path: ${codePath.trim()}`,
      ];

      if (selectedSecrets.length > 0) {
        lines.push(
          `create_secrets_from_connectors: ${selectedSecrets.join(", ")}`,
        );
      }

      await createSource(runtimeClient, modelName.trim(), lines.join("\n"));

      const newFilePath = getFilePathFromNameAndType(
        modelName.trim(),
        EntityType.Table,
      );
      await goto(`/files${newFilePath}`);
      overlay.set(null);
      onClose();
    } catch (err) {
      createError = err instanceof Error ? err.message : String(err);
    } finally {
      creating = false;
    }
  }

  function inferModelName(path: string) {
    if (!path || modelName) return;
    const slug = path.split("/").pop()?.replace(/\.py$/, "") ?? "";
    if (slug) modelName = slug.replace(/[^a-zA-Z0-9_]/g, "_");
  }

  function removePackage(pkg: string) {
    for (let i = 0; i < packageTemplates.length; i++) {
      if (
        packageTemplates[i].checked &&
        packageTemplates[i].packages.includes(pkg)
      ) {
        packageTemplates[i].checked = false;
        return;
      }
    }
    const customs = customPackageInput
      .split(",")
      .map((p) => p.trim())
      .filter((p) => p && p !== pkg);
    customPackageInput = customs.join(", ");
  }

  $: if (codePath && sourceMode === "custom") inferModelName(codePath);
</script>

<div class="flex flex-col h-full">
  <div class="flex-1 overflow-y-auto p-6">
    <!-- ── DETECT STEP ── -->
    {#if wizardStep === "detect"}
      <div class="flex flex-col gap-4">
        <h3 class="text-sm font-semibold text-fg-primary">Detecting Python</h3>
        {#if detecting}
          <div class="flex items-center gap-2 text-sm text-fg-secondary">
            <LoadingSpinner size="16px" />
            <span>Searching for Python installation...</span>
          </div>
        {:else if detectError}
          <div
            class="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800"
          >
            {detectError}
          </div>
        {:else if !pythonFound}
          <div class="p-3 bg-amber-50 border border-amber-200 rounded-md">
            <p class="text-sm font-medium text-amber-900">Python not found</p>
            <p class="text-sm text-amber-700 mt-1">
              Install Python 3.9+ and try again.
              <a
                href="https://github.com/pyenv/pyenv#installation"
                target="_blank"
                rel="noopener noreferrer"
                class="underline">Recommended: pyenv</a
              >
            </p>
          </div>
          <Button onClick={detectPython} type="secondary">
            Retry detection
          </Button>
        {/if}
      </div>
    {/if}

    <!-- ── SETUP STEP ── -->
    {#if wizardStep === "setup"}
      <div class="flex flex-col gap-4">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-green-500"></div>
          <span class="text-sm text-fg-secondary">
            Python {pythonVersion} found at {pythonPath}
          </span>
        </div>

        <h3 class="text-sm font-semibold text-fg-primary">
          Set up Python environment
        </h3>
        <p class="text-sm text-fg-secondary">
          Select package templates for your use case. Individual packages are
          shown below.
        </p>

        <!-- Package template grid -->
        <div class="grid grid-cols-2 gap-2">
          {#each packageTemplates as template, i}
            <button
              class="flex flex-col items-start gap-1 p-3 rounded-lg border text-left transition-colors
                {template.checked
                ? 'border-primary-500 bg-primary-50'
                : 'border-border hover:border-primary-300 bg-surface'}"
              on:click={() => {
                packageTemplates[i].checked = !packageTemplates[i].checked;
              }}
              disabled={settingUp}
            >
              <div class="flex items-center gap-2 w-full">
                <input
                  type="checkbox"
                  checked={template.checked}
                  class="pointer-events-none"
                  tabindex="-1"
                />
                <span class="text-sm font-medium text-fg-primary">
                  {template.label}
                </span>
              </div>
              <span class="text-xs text-fg-muted ml-6">
                {template.description}
              </span>
              <div class="flex flex-wrap gap-1 ml-6 mt-1">
                {#each template.packages as pkg}
                  <span
                    class="inline-block px-1.5 py-0.5 text-[10px] font-mono rounded
                      {template.checked
                      ? 'bg-primary-100 text-primary-700'
                      : 'bg-surface-muted text-fg-muted'}"
                  >
                    {pkg}
                  </span>
                {/each}
              </div>
            </button>
          {/each}
        </div>

        <!-- Custom packages input -->
        <div class="flex flex-col gap-1">
          <label
            for="custom-packages"
            class="text-sm font-medium text-fg-secondary"
          >
            Additional packages
          </label>
          <input
            id="custom-packages"
            type="text"
            bind:value={customPackageInput}
            placeholder="e.g. boto3, sqlalchemy, beautifulsoup4"
            disabled={settingUp}
            class="w-full px-3 py-2 border border-border rounded-md text-sm bg-surface text-fg-primary focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          />
        </div>

        <!-- Live package list -->
        <div class="flex flex-col gap-1.5">
          <div class="flex items-center justify-between">
            <span
              class="text-xs font-medium text-fg-muted uppercase tracking-wide"
            >
              Packages to install ({selectedPackages.length})
            </span>
            <span class="text-[10px] text-fg-muted">
              Synced to requirements.txt
            </span>
          </div>
          <div
            class="flex flex-wrap gap-1.5 p-3 rounded-md border border-border bg-surface-subtle min-h-[2.5rem]"
          >
            {#each selectedPackages as pkg}
              <span
                class="inline-flex items-center gap-1 px-2 py-1 text-xs font-mono rounded-md
                  {BASE_PACKAGES.includes(pkg)
                  ? 'bg-surface-muted text-fg-muted'
                  : 'bg-primary-50 text-primary-700 border border-primary-200'}"
              >
                {pkg}
                {#if !BASE_PACKAGES.includes(pkg)}
                  <button
                    class="ml-0.5 text-primary-400 hover:text-primary-600"
                    on:click|stopPropagation={() => removePackage(pkg)}
                    disabled={settingUp}
                    title="Remove {pkg}"
                  >
                    &times;
                  </button>
                {/if}
              </span>
            {/each}
          </div>
        </div>

        {#if setupError}
          <div
            class="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800"
          >
            {setupError}
          </div>
        {/if}
      </div>
    {/if}

    <!-- ── SOURCE STEP ── -->
    {#if wizardStep === "source"}
      <div class="flex flex-col gap-4">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-green-500"></div>
          <span class="text-sm text-fg-secondary">
            Python environment ready 
          </span>
        </div>

        <h3 class="text-sm font-semibold text-fg-primary">
          Configure Python source
        </h3>

        <!-- Tab toggle: Template / Custom -->
        <div
          class="inline-flex rounded-md border border-border bg-surface-subtle p-0.5 self-start"
        >
          <button
            class="px-3 py-1 text-sm rounded transition-colors
              {sourceMode === 'template'
              ? 'bg-surface text-fg-primary font-medium shadow-sm'
              : 'text-fg-muted hover:text-fg-secondary'}"
            on:click={() => {
              sourceMode = "template";
            }}
            disabled={creating}
          >
            Template
          </button>
          <button
            class="px-3 py-1 text-sm rounded transition-colors
              {sourceMode === 'custom'
              ? 'bg-surface text-fg-primary font-medium shadow-sm'
              : 'text-fg-muted hover:text-fg-secondary'}"
            on:click={() => {
              sourceMode = "custom";
              selectedTemplateId = null;
              codePath = "";
              modelName = "";
            }}
            disabled={creating}
          >
            Custom script
          </button>
        </div>

        <!-- Template grid -->
        {#if sourceMode === "template"}
          <div class="grid grid-cols-2 gap-2">
            {#each pythonTemplates as tmpl}
              <button
                class="flex flex-col items-start gap-0.5 p-3 rounded-lg border text-left transition-colors
                  {selectedTemplateId === tmpl.id
                  ? 'border-primary-500 bg-primary-50'
                  : 'border-border hover:border-primary-300 bg-surface'}"
                on:click={() => selectTemplate(tmpl.id)}
                disabled={creating}
              >
                <span class="text-sm font-medium text-fg-primary">
                  {tmpl.label}
                </span>
                <span class="text-xs text-fg-muted leading-tight">
                  {tmpl.description}
                </span>
                {#if selectedTemplateId === tmpl.id}
                  <span
                    class="mt-1 text-[10px] font-mono text-fg-muted bg-surface-muted px-1.5 py-0.5 rounded"
                  >
                    → {tmpl.defaultPath}
                  </span>
                {/if}
              </button>
            {/each}
          </div>
        {:else}
          <!-- Custom script path -->
          <div class="flex flex-col gap-1">
            <label
              for="code-path"
              class="text-sm font-medium text-fg-secondary"
            >
              Script path
            </label>
            <input
              id="code-path"
              type="text"
              bind:value={codePath}
              placeholder="scripts/extract.py"
              disabled={creating}
              class="w-full px-3 py-2 border border-border rounded-md text-sm font-mono bg-surface text-fg-primary focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
            />
            <span class="text-xs text-fg-muted">
              Path to an existing Python script relative to the project root
            </span>
          </div>
        {/if}

        <!-- Model name -->
        <div class="flex flex-col gap-1">
          <label
            for="model-name"
            class="text-sm font-medium text-fg-secondary"
          >
            Model name
          </label>
          <input
            id="model-name"
            type="text"
            bind:value={modelName}
            placeholder="my_python_model"
            disabled={creating}
            class="w-full px-3 py-2 border border-border rounded-md text-sm bg-surface text-fg-primary focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          />
        </div>

        <!-- Connector secrets dropdown -->
        <div class="flex flex-col gap-1.5">
          <label
            for="secrets-select"
            class="text-sm font-medium text-fg-secondary"
          >
            Pass secrets from connectors
          </label>
          {#if $connectorsQuery.isLoading}
            <div class="flex items-center gap-2 text-xs text-fg-muted">
              <LoadingSpinner size="12px" />
              Loading connectors...
            </div>
          {:else if $connectorsQuery.data && $connectorsQuery.data.length > 0}
            <select
              id="secrets-select"
              class="w-full px-3 py-2 border border-border rounded-md text-sm bg-surface text-fg-primary focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              on:change={(e) => {
                const val = e.currentTarget.value;
                if (val && !selectedSecrets.includes(val)) {
                  selectedSecrets = [...selectedSecrets, val];
                }
                e.currentTarget.value = "";
              }}
              disabled={creating}
              value=""
            >
              <option value="" disabled>Select a connector...</option>
              {#each $connectorsQuery.data as connector}
                {@const name = connector.meta?.name?.name ?? ""}
                {@const driver = connector.connector?.spec?.driver ?? ""}
                {#if name && !selectedSecrets.includes(name)}
                  <option value={name}>{name} ({driver})</option>
                {/if}
              {/each}
            </select>
            {#if selectedSecrets.length > 0}
              <div class="flex flex-wrap gap-1.5 mt-1">
                {#each selectedSecrets as secret}
                  <span
                    class="inline-flex items-center gap-1 px-2 py-1 text-xs font-mono rounded-md bg-primary-50 text-primary-700 border border-primary-200"
                  >
                    {secret}
                    <button
                      class="ml-0.5 text-primary-400 hover:text-primary-600"
                      on:click={() => toggleSecret(secret)}
                      disabled={creating}
                    >
                      &times;
                    </button>
                  </span>
                {/each}
              </div>
            {/if}
          {:else}
            <span class="text-xs text-fg-muted italic">
              No connectors configured yet
            </span>
          {/if}
        </div>

        {#if createError}
          <div
            class="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800"
          >
            {createError}
          </div>
        {/if}
      </div>
    {/if}
  </div>

  <!-- ── FOOTER ── -->
  <div
    class="w-full bg-surface-subtle border-t border-border p-6 flex justify-between gap-2"
  >
    <Button onClick={onBack} type="secondary">Back</Button>

    <div class="flex gap-2">
      {#if wizardStep === "setup"}
        <Button
          onClick={() => {
            wizardStep = "source";
          }}
          type="secondary"
        >
          Skip setup
        </Button>
        <Button
          onClick={setupEnvironment}
          loading={settingUp}
          loadingCopy="Installing {selectedPackages.length} packages..."
          type="primary"
          disabled={settingUp}
        >
          Install {selectedPackages.length} packages
        </Button>
      {:else if wizardStep === "source"}
        <Button
          onClick={createModel}
          loading={creating}
          loadingCopy="Creating..."
          type="primary"
          disabled={creating ||
            !modelName.trim() ||
            (sourceMode === "template" && !selectedTemplateId) ||
            (sourceMode === "custom" && !codePath.trim())}
        >
          Create model
        </Button>
      {/if}
    </div>
  </div>
</div>
