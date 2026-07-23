<script lang="ts">
  import { autocompletion } from "@codemirror/autocomplete";
  import { keywordCompletionSource, sql } from "@codemirror/lang-sql";
  import {
    EditorState,
    Prec,
    StateEffect,
    StateField,
  } from "@codemirror/state";
  import {
    Decoration,
    EditorView,
    keymap,
    type DecorationSet,
  } from "@codemirror/view";
  import { onMount } from "svelte";
  import { base as baseExtensions } from "@rilldata/web-common/components/editor/presets/base";
  import { DuckDBSQL } from "@rilldata/web-common/components/editor/presets/duckDBDialect";
  import {
    getStatementRange,
    trimRange,
    type SqlExecution,
  } from "./query-utils";

  let {
    initialValue = "",
    disabled = false,
    onchange = (_value: string) => {},
    onrun = (_execution: SqlExecution) => {},
  }: {
    initialValue?: string;
    disabled?: boolean;
    onchange?: (value: string) => void;
    onrun?: (execution: SqlExecution) => void;
  } = $props();

  let parent: HTMLDivElement;
  let editor: EditorView | null = null;

  const setExecutedRange = StateEffect.define<
    { from: number; to: number } | undefined
  >();
  const executedRange = StateField.define<DecorationSet>({
    create: () => Decoration.none,
    update: (decorations, transaction) => {
      decorations = decorations.map(transaction.changes);
      for (const effect of transaction.effects) {
        if (!effect.is(setExecutedRange)) continue;
        decorations = effect.value
          ? Decoration.set([
              Decoration.mark({ class: "cm-executed-statement" }).range(
                effect.value.from,
                effect.value.to,
              ),
            ])
          : Decoration.none;
      }
      return decorations;
    },
    provide: (field) => EditorView.decorations.from(field),
  });

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: initialValue,
        extensions: [
          baseExtensions(),
          sql({ dialect: DuckDBSQL }),
          autocompletion({
            override: [keywordCompletionSource(DuckDBSQL)],
            icons: false,
          }),
          executedRange,
          Prec.highest(
            keymap.of([
              {
                key: "Mod-Enter",
                run: () => {
                  runCurrent();
                  return true;
                },
              },
            ]),
          ),
          EditorView.contentAttributes.of({
            "aria-label": "SQL worksheet editor",
          }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) onchange(update.state.doc.toString());
          }),
        ],
      }),
      parent,
    });

    return () => editor?.destroy();
  });

  export function runCurrent() {
    if (!editor || disabled) return;

    const selection = editor.state.selection.main;
    const document = editor.state.doc.toString();
    const range = selection.empty
      ? getStatementRange(document, selection.head)
      : trimRange(document, { from: selection.from, to: selection.to });
    if (!range) return;

    const endPosition = Math.max(range.from, range.to - 1);
    editor.dispatch({
      effects: setExecutedRange.of(range),
    });
    onrun({
      ...range,
      sql: editor.state.sliceDoc(range.from, range.to),
      startLine: editor.state.doc.lineAt(range.from).number,
      endLine: editor.state.doc.lineAt(endPosition).number,
    });
  }
</script>

<div bind:this={parent} class="size-full overflow-hidden"></div>

<style lang="postcss">
  div :global(.cm-editor) {
    @apply h-full;
    padding-top: 2px;
    font-size: clamp(13px, calc(10px + 0.3vw), 15px);
  }

  div :global(.cm-scroller) {
    @apply overflow-auto;
  }

  div :global(.cm-executed-statement) {
    background-color: color-mix(
      in oklab,
      var(--color-primary-500) 10%,
      transparent
    );
  }
</style>
