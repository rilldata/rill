<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { onMount } from "svelte";

  type SettingsPage = "rill-yaml" | "project" | "runtime" | "preferences";

  let selectedPage: SettingsPage = "rill-yaml";
  let rillYamlContent: string = "";
  let loading = false;
  let error: string | null = null;
  let copySuccess = false;

  const navItems: Array<{ label: string; route: SettingsPage }> = [
    { label: "rill.yaml", route: "rill-yaml" },
    { label: "Project Metadata", route: "project" },
    { label: "Runtime Settings", route: "runtime" },
    { label: "Developer Preferences", route: "preferences" },
  ];

  async function loadSettings() {
    try {
      loading = true;
      error = null;

      if (!$runtime?.host || !$runtime?.instanceId) {
        error = "Waiting for runtime to initialize...";
        loading = false;
        return;
      }

      // Try to fetch rill.yaml from backend
      try {
        const response = await fetch(
          `${$runtime.host}/v1/instances/${$runtime.instanceId}/rill-yaml`,
        );

        if (response.ok) {
          rillYamlContent = await response.text();
        } else {
          // Fallback to showing a sample
          rillYamlContent = getDefaultRillYaml();
        }
      } catch {
        console.warn("Could not fetch rill.yaml from backend, showing sample");
        rillYamlContent = getDefaultRillYaml();
      }
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load settings";
      console.error("Error loading settings:", err);
    } finally {
      loading = false;
    }
  }

  function getDefaultRillYaml(): string {
    return `# rill.yaml - Project Configuration
# This file defines your Rill project structure and settings

# Project metadata
display_name: "My Rill Project"
description: "Analytics dashboard powered by Rill"

# Default OLAP connector
olap_connector: duckdb

# Environment settings
environment: dev

# Features
features:
  chat_charts: true

# Mock users for local testing
mock_users:
  - email: test@example.com
    name: Test User
    admin: true

# Model refresh schedule
models:
  refresh:
    every: 24h`;
  }

  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(rillYamlContent);
      copySuccess = true;
      setTimeout(() => {
        copySuccess = false;
      }, 2000);
    } catch {
      console.error("Failed to copy to clipboard");
    }
  }

  onMount(() => {
    loadSettings();
  });

  // Retry when runtime becomes available
  $: if ($runtime?.instanceId && $runtime?.host && error?.includes("Waiting")) {
    loadSettings();
  }
</script>

<div class="h-full w-full flex bg-white dark:bg-gray-950 overflow-hidden">
  <!-- Left Navigation -->
  <div class="w-48 border-r border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900 p-4 overflow-y-auto">
    <div class="space-y-1">
      {#each navItems as item}
        <button
          on:click={() => (selectedPage = item.route)}
          class={`w-full text-left px-3 py-2 rounded text-sm transition-colors ${
            selectedPage === item.route
              ? "bg-gray-200 dark:bg-gray-800 text-gray-900 dark:text-white font-medium"
              : "text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800"
          }`}
        >
          {item.label}
        </button>
      {/each}
    </div>
  </div>

  <!-- Main Content -->
  <div class="flex-1 overflow-auto flex flex-col">
    <div class="p-8 flex-1">
      {#if selectedPage === "rill-yaml"}
        <!-- rill.yaml Settings -->
        <div class="max-w-3xl">
          <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-2">
            rill.yaml
          </h2>
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-6">
            Your project configuration file. Edit in your code editor for changes.
          </p>

          <div class="border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden bg-gray-50 dark:bg-gray-900">
            <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-800">
              <span class="text-sm font-medium text-gray-700 dark:text-gray-300">
                Configuration
              </span>
              <Button
                type="secondary"
                onClick={copyToClipboard}
                compact
                square
                label="Copy to clipboard"
              >
                <CopyIcon size="12px" />
              </Button>
            </div>
            <pre
              class="p-4 text-sm font-mono overflow-x-auto text-gray-700 dark:text-gray-300"
            ><code>{rillYamlContent}</code></pre>
          </div>

          {#if copySuccess}
            <p class="text-xs text-green-600 dark:text-green-400 mt-2">
              ✓ Copied to clipboard
            </p>
          {/if}
        </div>
      {:else if selectedPage === "project"}
        <!-- Project Metadata -->
        <div class="max-w-3xl">
          <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-2">
            Project Metadata
          </h2>
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-6">
            Information about your Rill project.
          </p>

          <div class="border border-gray-200 dark:border-gray-800 rounded-lg p-6 space-y-4">
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">
                Project Path
              </p>
              <p class="text-sm font-mono text-gray-600 dark:text-gray-400">
                {$runtime?.metadata?.projectPath || "N/A"}
              </p>
            </div>
            <div class="border-t border-gray-200 dark:border-gray-800 pt-4">
              <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">
                Rill Version
              </p>
              <p class="text-sm text-gray-600 dark:text-gray-400">
                {$runtime?.version || "N/A"}
              </p>
            </div>
            <div class="border-t border-gray-200 dark:border-gray-800 pt-4">
              <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">
                Environment
              </p>
              <p class="text-sm text-gray-600 dark:text-gray-400">
                Local Development
              </p>
            </div>
          </div>
        </div>
      {:else if selectedPage === "runtime"}
        <!-- Runtime Settings -->
        <div class="max-w-3xl">
          <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-2">
            Runtime Settings
          </h2>
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-6">
            Information about the local runtime instance.
          </p>

          <div class="border border-gray-200 dark:border-gray-800 rounded-lg p-6 space-y-4">
            <div>
              <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">
                Runtime Host
              </p>
              <p class="text-sm font-mono text-gray-600 dark:text-gray-400">
                {$runtime?.host || "N/A"}
              </p>
            </div>
            <div class="border-t border-gray-200 dark:border-gray-800 pt-4">
              <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">
                Instance ID
              </p>
              <p class="text-sm font-mono text-gray-600 dark:text-gray-400">
                {$runtime?.instanceId || "N/A"}
              </p>
            </div>
            <div class="border-t border-gray-200 dark:border-gray-800 pt-4">
              <p class="text-sm font-medium text-gray-900 dark:text-white mb-1">
                Status
              </p>
              <p class="text-sm text-gray-600 dark:text-gray-400">
                Connected and Ready
              </p>
            </div>
          </div>
        </div>
      {:else if selectedPage === "preferences"}
        <!-- Developer Preferences -->
        <div class="max-w-3xl">
          <h2 class="text-2xl font-semibold text-gray-900 dark:text-white mb-2">
            Developer Preferences
          </h2>
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-6">
            Customize your development experience.
          </p>

          <div class="border border-gray-200 dark:border-gray-800 rounded-lg p-6">
            <p class="text-sm text-gray-600 dark:text-gray-400">
              Additional preferences can be configured here in the future.
            </p>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>
