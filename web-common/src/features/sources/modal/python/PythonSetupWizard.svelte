<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import LoadingSpinner from "@rilldata/web-common/components/icons/LoadingSpinner.svelte";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { useRuntimeClient } from "../../../../runtime-client/v2";
  import {
    runtimeServiceDetectPython,
    runtimeServiceGetFile,
    runtimeServiceSetupPythonEnvironment,
  } from "../../../../runtime-client/v2/gen/runtime-service";
  import { createSource } from "../createSource";

  export let onClose: () => void = () => {};
  export let onBack: () => void = () => {};

  const runtimeClient = useRuntimeClient();

  // Wizard step: "detect" → "setup" → "source"
  let wizardStep: "detect" | "setup" | "source" = "detect";

  // Detect state
  let detecting = true;
  let pythonFound = false;
  let pythonPath = "";
  let pythonVersion = "";
  let detectError = "";

  // Setup state
  let settingUp = false;
  let setupComplete = false;
  let setupError = "";
  let installedPackages: string[] = [];

  // Package templates with their pip packages
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

  // Base packages always included
  const BASE_PACKAGES = ["pandas", "pyarrow"];

  // Reactive: build the full package list from selections + custom input
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

  // Source step state
  let codePath = "";
  let modelName = "";
  let creating = false;
  let createError = "";

  // Run detection + load existing requirements on mount
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
          (line: string) => line && !line.startsWith("#") && !line.startsWith("-"),
        );

      // Check matching templates
      for (let i = 0; i < packageTemplates.length; i++) {
        const allPresent = packageTemplates[i].packages.every((p) =>
          existingPkgs.some(
            (ep: string) => ep === p || ep.startsWith(p + "==") || ep.startsWith(p + ">="),
          ),
        );
        if (allPresent) {
          packageTemplates[i].checked = true;
        }
      }

      // Add any packages not covered by templates to custom input
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
      // No requirements.txt yet; that's fine
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
    if (!codePath.trim() || !modelName.trim()) return;
    creating = true;
    createError = "";
    try {
      const yaml = [
        "# Model YAML",
        "# Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/python",
        "",
        "type: model",
        "materialize: true",
        "",
        "connector: python",
        "",
        `code_path: ${codePath.trim()}`,
      ].join("\n");

      await createSource(runtimeClient, modelName.trim(), yaml);
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
    if (slug) {
      modelName = slug.replace(/[^a-zA-Z0-9_]/g, "_");
    }
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

  $: if (codePath) inferModelName(codePath);
</script>

<div class="flex flex-col h-full">
  <div class="flex-1 overflow-y-auto p-6">
    <!-- DETECT STEP -->
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
                class="underline"
              >
                Recommended: pyenv
              </a>
            </p>
          </div>
          <Button onClick={detectPython} type="secondary">
            Retry detection
          </Button>
        {/if}
      </div>
    {/if}

    <!-- SETUP STEP -->
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

        <!-- Template multi-select grid -->
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

    <!-- SOURCE STEP -->
    {#if wizardStep === "source"}
      <div class="flex flex-col gap-4">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-green-500"></div>
          <span class="text-sm text-fg-secondary">
            Python environment ready ({installedPackages.length} packages installed)
          </span>
        </div>

        <h3 class="text-sm font-semibold text-fg-primary">
          Configure Python source
        </h3>
        <p class="text-sm text-fg-secondary">
          Point to a Python script that writes a Parquet file to the
          <code class="bg-surface-muted px-1 rounded text-xs">RILL_OUTPUT_PATH</code>
          environment variable.
        </p>

        <div class="flex flex-col gap-1">
          <label for="code-path" class="text-sm font-medium text-fg-secondary">
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
          <span class="text-xs text-fg-muted">Relative to project root</span>
        </div>

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

  <!-- FOOTER -->
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
          disabled={creating || !codePath.trim() || !modelName.trim()}
        >
          Create model
        </Button>
      {/if}
    </div>
  </div>
</div>
