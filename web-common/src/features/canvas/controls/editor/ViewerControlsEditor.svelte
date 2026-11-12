<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { ArrowDown, ArrowUp, Pencil, Plus, Trash2 } from "lucide-svelte";
  import { applyCanvasEdit } from "../../inspector/BlockEditor.svelte";
  import type { YAMLPath } from "../../layout-util";
  import {
    addControl,
    draftFromYAML,
    moveControl,
    moveControlToLevel,
    readControls,
    removeControl,
    updateControl,
    type ControlDraft,
    type ControlLevelRef,
  } from "../control-edit";
  import ControlDialog from "./ControlDialog.svelte";
  import { controlRequest } from "./control-request";
  import { controlKindLabel } from "./labels";

  // The "Viewer controls" section of the inspector: the controls declared on one level
  // (the canvas, a container, a tab group, a tab or a widget), with add, edit, reorder and remove.
  let {
    canvas,
    path,
    hint,
    levels = undefined,
    metricsViewName = undefined,
  }: {
    canvas: CanvasEntity;
    // The YAML path of the level, e.g. [] for the canvas or ["rows", 2] for a container.
    path: YAMLPath;
    // Which widgets the level's controls change.
    hint: string;
    // For a widget: the levels its controls can be declared on, so a control can be moved
    // to one of them. Without it, controls stay on this level.
    levels?: ControlLevelRef[];
    // For a widget: its metrics view, whose fields the dialog offers.
    metricsViewName?: string;
  } = $props();

  // The dialog's target: a new control (optionally prefilled) or the index of an existing one.
  let dialog = $state<
    { index: number | null; draft: ControlDraft | undefined } | undefined
  >(undefined);

  let parsedContent = $derived(canvas.parsedContent);
  let metricsViews = $derived(canvas._metricsViews);
  let controls = $derived(readControls($parsedContent, path));
  let pathKey = $derived(path.join("/"));
  let dialogLevels = $derived(
    levels ?? [{ kind: "canvas" as const, path, name: "" }],
  );

  // A "Let viewers choose" link next to one of the widget's fields asks for the dialog.
  $effect(() => {
    const request = $controlRequest;
    if (request && request.itemPath === pathKey) {
      dialog = { index: null, draft: request.draft };
      controlRequest.set(null);
    }
  });

  let fieldLabels = $derived.by(() => {
    const labels = new Map<string, string>();
    for (const view of Object.values($metricsViews)) {
      const spec = view?.state?.validSpec;
      for (const f of [
        ...(spec?.measures ?? []),
        ...(spec?.dimensions ?? []),
      ]) {
        if (f.name && !labels.has(f.name)) {
          labels.set(f.name, f.displayName || f.name);
        }
      }
    }
    return labels;
  });

  function describe(raw: Record<string, unknown>) {
    const draft = draftFromYAML(raw);
    const kind = controlKindLabel(draft?.kind ?? "slot");
    const target =
      draft?.field || (typeof raw.slot === "string" ? raw.slot : "");
    const title =
      (typeof raw.label === "string" && raw.label) ||
      (target ? (fieldLabels.get(target) ?? target) : kind);
    const count = Array.isArray(raw.options) ? raw.options.length : 0;
    const detail = count
      ? m.canvas_viewer_controls_summary({ kind, count })
      : kind;
    return { title, detail, editable: draft !== null, draft };
  }

  function edit(mutate: Parameters<typeof applyCanvasEdit>[2]) {
    if (!canvas.fileArtifact) return;
    return applyCanvasEdit(canvas.fileArtifact, true, mutate);
  }

  async function save(draft: ControlDraft, levelPath: YAMLPath) {
    const index = dialog?.index ?? null;
    dialog = undefined;
    await edit((doc) => {
      if (index === null) {
        addControl(doc, levelPath, draft);
      } else if (levelPath.join("/") === pathKey) {
        updateControl(doc, path, index, draft);
      } else {
        moveControlToLevel(doc, path, index, levelPath, draft);
      }
    });
  }
</script>

<div class="viewer-controls">
  <div class="flex flex-col gap-y-0.5">
    <span class="text-xs font-medium text-fg-secondary">
      {m.canvas_viewer_controls()}
    </span>
    <span class="text-[11px] text-fg-secondary">{hint}</span>
  </div>

  {#if controls.length}
    <ul class="flex flex-col gap-y-1.5" aria-label={m.canvas_viewer_controls()}>
      {#each controls as raw, index (index)}
        {@const info = describe(raw)}
        <li class="control-row">
          <div class="flex min-w-0 flex-1 flex-col">
            <span class="truncate text-xs font-medium">{info.title}</span>
            <span class="truncate text-[11px] text-fg-secondary">
              {info.editable
                ? info.detail
                : m.canvas_viewer_controls_yaml_only({ kind: info.detail })}
            </span>
          </div>
          <div class="actions">
            <button
              type="button"
              title={m.canvas_viewer_controls_move_up()}
              aria-label={m.canvas_viewer_controls_move_up()}
              disabled={index === 0}
              onclick={() => edit((doc) => moveControl(doc, path, index, -1))}
            >
              <ArrowUp size="14px" />
            </button>
            <button
              type="button"
              title={m.canvas_viewer_controls_move_down()}
              aria-label={m.canvas_viewer_controls_move_down()}
              disabled={index === controls.length - 1}
              onclick={() => edit((doc) => moveControl(doc, path, index, 1))}
            >
              <ArrowDown size="14px" />
            </button>
            <button
              type="button"
              title={m.canvas_viewer_controls_edit()}
              aria-label={m.canvas_viewer_controls_edit()}
              disabled={!info.editable}
              onclick={() =>
                (dialog = { index, draft: info.draft ?? undefined })}
            >
              <Pencil size="14px" />
            </button>
            <button
              type="button"
              title={m.canvas_viewer_controls_remove()}
              aria-label={m.canvas_viewer_controls_remove()}
              onclick={() => edit((doc) => removeControl(doc, path, index))}
            >
              <Trash2 size="14px" />
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}

  <Button
    type="secondary"
    onClick={() => (dialog = { index: null, draft: undefined })}
  >
    <Plus size="14px" />
    {m.canvas_viewer_controls_add()}
  </Button>
</div>

{#if dialog}
  <ControlDialog
    {canvas}
    levels={dialogLevels}
    levelPath={path}
    initial={dialog.draft}
    isNew={dialog.index === null}
    {metricsViewName}
    onSave={save}
    onCancel={() => (dialog = undefined)}
  />
{/if}

<style lang="postcss">
  .viewer-controls {
    @apply flex flex-col gap-y-2 px-5 py-3 border-t;
  }

  .control-row {
    @apply flex items-center gap-x-2 rounded-md border px-2 py-1.5;
  }

  .actions {
    @apply flex flex-none items-center gap-x-0.5;
  }

  .actions button {
    @apply grid size-6 place-content-center rounded text-fg-secondary;
    @apply hover:bg-surface-subtle hover:text-fg-primary;
    @apply disabled:pointer-events-none disabled:opacity-30;
  }
</style>
