<script lang="ts">
  import { replaceState } from "$app/navigation";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { DASHBOARD_WIDTH } from "@rilldata/web-common/features/canvas/constants";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import ComponentOptions from "@rilldata/web-common/features/dashboards/canvas/ComponentOptions.svelte";
  import Inspector from "@rilldata/web-common/layout/workspace/Inspector.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { parseDocument, YAMLMap } from "yaml";
  import {
    ResourceKind,
    useFilteredResources,
  } from "../entity-management/resource-selectors";
  import SidebarWrapper from "../visual-editing/SidebarWrapper.svelte";
  import ThemeInput from "../visual-editing/ThemeInput.svelte";

  export let viewingDashboard: boolean;
  export let switchView: () => void;

  const { fileArtifact, validSpecStore } = getCanvasStateManagers();

  $: ({ instanceId } = $runtime);
  $: ({ localContent, remoteContent, saveContent, path } = $fileArtifact);

  $: canvasSpec = $validSpecStore;

  $: parsedDocument = parseDocument($localContent ?? $remoteContent ?? "");

  $: rawTitle = parsedDocument.get("title");
  $: rawDisplayName = parsedDocument.get("display_name");
  $: rawTheme = parsedDocument.get("theme"); //Add property
  $: maxWidth = DASHBOARD_WIDTH; //Add property

  $: title = stringGuard(rawTitle) || stringGuard(rawDisplayName);

  $: themesQuery = useFilteredResources(instanceId, ResourceKind.Theme);

  $: themeNames = ($themesQuery?.data ?? [])
    .map((theme) => theme.meta?.name?.name ?? "")
    .filter((string) => !string.endsWith("--theme"));

  $: theme = !rawTheme
    ? undefined
    : typeof rawTheme === "string"
      ? rawTheme
      : rawTheme instanceof YAMLMap
        ? canvasSpec?.embeddedTheme
        : undefined;

  function stringGuard(value: unknown | undefined): string {
    return value && typeof value === "string" ? value : "";
  }

  async function updateProperties(
    newRecord: Record<string, unknown>,
    removeProperties?: Array<string | string[]>,
  ) {
    Object.entries(newRecord).forEach(([property, value]) => {
      if (!value) {
        parsedDocument.delete(property);
      } else {
        parsedDocument.set(property, value);
      }
    });

    if (removeProperties) {
      removeProperties.forEach((prop) => {
        try {
          if (Array.isArray(prop)) {
            parsedDocument.deleteIn(prop);
          } else {
            parsedDocument.delete(prop);
          }
        } catch {
          // ignore
        }
      });
    }

    killState();

    await saveContent(parsedDocument.toString());
  }

  function killState() {
    replaceState(window.location.origin + window.location.pathname, {});
  }
</script>

<Inspector filePath={path}>
  <SidebarWrapper title="Edit dashboard">
    <p class="text-slate-500 text-sm">Changes below will be auto-saved.</p>

    <Input
      hint="Shown in global header and when deployed to Rill Cloud"
      capitalizeLabel={false}
      textClass="text-sm"
      label="Display name"
      bind:value={title}
      onBlur={async () => {
        await updateProperties({ display_name: title }, ["title"]);
      }}
      onEnter={async () => {
        await updateProperties({ display_name: title });
      }}
    />

    <ComponentOptions
      on:select={(e) => {
        console.log(e);
      }}
    />

    <!-- TODO: Support number input -->
    <Input
      hint="Max width for the canvas"
      capitalizeLabel={false}
      textClass="text-sm"
      label="Max width"
      inputType="number"
      bind:value={maxWidth}
      onBlur={async () => {
        await updateProperties({ display_name: title }, ["title"]);
      }}
      onEnter={async () => {
        await updateProperties({ display_name: title });
      }}
    />

    <ThemeInput
      {theme}
      {themeNames}
      onThemeChange={async (value) => {
        if (!value) {
          await updateProperties({}, ["theme"]);
        } else {
          await updateProperties({ theme: value });
        }
      }}
      onColorChange={async (primary, secondary) => {
        await updateProperties({
          theme: {
            colors: {
              primary,
              secondary,
            },
          },
        });
      }}
    />

    <svelte:fragment slot="footer">
      {#if viewingDashboard}
        <footer
          class="flex flex-col gap-y-4 mt-auto border-t px-5 py-5 pb-6 w-full text-sm text-gray-500"
        >
          <p>
            For more options,
            <button on:click={switchView} class="text-primary-600 font-medium">
              edit in YAML
            </button>
          </p>
        </footer>
      {/if}
    </svelte:fragment>
  </SidebarWrapper>
</Inspector>

<style lang="postcss">
</style>
