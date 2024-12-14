<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import { DASHBOARD_WIDTH } from "@rilldata/web-common/features/canvas/constants";
  import { getParsedDocument } from "@rilldata/web-common/features/canvas/inspector/selectors";
  import { getCanvasStateManagers } from "@rilldata/web-common/features/canvas/state-managers/state-managers";
  import {
    ResourceKind,
    useFilteredResources,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import ThemeInput from "@rilldata/web-common/features/visual-editing/ThemeInput.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { YAMLMap } from "yaml";

  export let updateProperties: (
    newRecord: Record<string, unknown>,
    removeProperties?: Array<string | string[]>,
  ) => Promise<void>;
  const { fileArtifact, validSpecStore } = getCanvasStateManagers();

  $: ({ instanceId } = $runtime);

  $: parsedDocument = getParsedDocument($fileArtifact);

  $: rawTitle = $parsedDocument.get("title");
  $: rawDisplayName = $parsedDocument.get("display_name");
  $: rawTheme = $parsedDocument.get("theme");
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
        ? $validSpecStore?.embeddedTheme
        : undefined;

  function stringGuard(value: unknown | undefined): string {
    return value && typeof value === "string" ? value : "";
  }
</script>

<SidebarWrapper title="Edit page">
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
</SidebarWrapper>
