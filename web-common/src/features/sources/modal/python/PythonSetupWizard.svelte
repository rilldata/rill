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

  // Package selection
  const packageSets = [
    { name: "stripe", label: "Stripe", description: "Stripe API (billing, charges, subscriptions)", packages: ["stripe"], checked: false },
    { name: "google-analytics", label: "Google Analytics", description: "Google Analytics Data API (GA4)", packages: ["google-analytics-data"], checked: false },
    { name: "dbt", label: "dbt", description: "dbt Core with DuckDB adapter", packages: ["dbt-core", "dbt-duckdb"], checked: false },
    { name: "requests", label: "Requests", description: "HTTP requests library (general REST APIs)", packages: ["requests"], checked: false },
  ];
  let customPackages = "";

  // Source step state
  let codePath = "";
  let modelName = "";
  let creating = false;
  let createError = "";

  // Run detection on mount
  detectPython();

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

  async function setupEnvironment() {
    settingUp = true;
    setupError = "";
    try {
      // Collect selected packages
      const selectedPackages: string[] = [];
      for (const ps of packageSets) {
        if (ps.checked) {
          selectedPackages.push(...ps.packages);
        }
      }
      if (customPackages.trim()) {
        for (const pkg of customPackages.split(",")) {
          const trimmed = pkg.trim();
          if (trimmed) selectedPackages.push(trimmed);
        }
      }

      const result = await runtimeServiceSetupPythonEnvironment(runtimeClient, {
        packages: selectedPackages,
        pythonPath: pythonPath || undefined,
      });

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
          <div class="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800">
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
          <Button onClick={detectPython} type="secondary">Retry detection</Button>
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
          A virtual environment will be created with pandas and pyarrow. Select additional packages:
        </p>

        <div class="flex flex-col gap-2">
          {#each packageSets as ps}
            <label class="flex items-start gap-2 cursor-pointer">
              <input
                type="checkbox"
                bind:checked={ps.checked}
                class="mt-0.5"
                disabled={settingUp}
              />
              <div>
                <span class="text-sm font-medium text-fg-primary">{ps.label}</span>
                <span class="text-sm text-fg-secondary"> — {ps.description}</span>
              </div>
            </label>
          {/each}
        </div>

        <div class="flex flex-col gap-1">
          <label for="custom-packages" class="text-sm font-medium text-fg-primary">
            Additional packages
          </label>
          <input
            id="custom-packages"
            type="text"
            bind:value={customPackages}
            placeholder="comma-separated, e.g. boto3, sqlalchemy"
            disabled={settingUp}
            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          />
        </div>

        {#if setupError}
          <div class="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800">
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
          <code class="bg-gray-100 px-1 rounded text-xs">RILL_OUTPUT_PATH</code>
          environment variable.
        </p>

        <div class="flex flex-col gap-1">
          <label for="code-path" class="text-sm font-medium text-fg-primary">
            Script path
          </label>
          <input
            id="code-path"
            type="text"
            bind:value={codePath}
            placeholder="scripts/extract.py"
            disabled={creating}
            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          />
          <span class="text-xs text-fg-secondary">Relative to project root</span>
        </div>

        <div class="flex flex-col gap-1">
          <label for="model-name" class="text-sm font-medium text-fg-primary">
            Model name
          </label>
          <input
            id="model-name"
            type="text"
            bind:value={modelName}
            placeholder="my_python_model"
            disabled={creating}
            class="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
          />
        </div>

        {#if createError}
          <div class="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-800">
            {createError}
          </div>
        {/if}
      </div>
    {/if}
  </div>

  <!-- FOOTER -->
  <div class="w-full bg-surface-subtle border-t border-gray-200 p-6 flex justify-between gap-2">
    <Button onClick={onBack} type="secondary">Back</Button>

    <div class="flex gap-2">
      {#if wizardStep === "setup"}
        <Button onClick={() => { wizardStep = "source"; }} type="secondary">
          Skip setup
        </Button>
        <Button
          onClick={setupEnvironment}
          loading={settingUp}
          loadingCopy="Setting up..."
          type="primary"
          disabled={settingUp}
        >
          Set up environment
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
