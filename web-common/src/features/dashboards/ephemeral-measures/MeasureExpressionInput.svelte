<!-- @component
The expression editor of the ephemeral measure editors.

A single-line tiptap editor. Typing "@" opens a picker of the measures the
expression may reference, searchable by display name as well as name, and a
picked measure is shown as a chip with its display name (like a mention in the
AI chat). `value` is the plain expression: chips serialize to the measure's
field name (see measure-mention.ts).
-->
<script lang="ts">
  import * as Kbd from "@rilldata/web-common/components/kbd";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
  import {
    autoUpdate,
    computePosition,
    flip,
    offset,
    shift,
  } from "@floating-ui/dom";
  import { Editor, Extension } from "@tiptap/core";
  import Document from "@tiptap/extension-document";
  import Mention from "@tiptap/extension-mention";
  import Paragraph from "@tiptap/extension-paragraph";
  import Text from "@tiptap/extension-text";
  import { Placeholder } from "@tiptap/extensions";
  import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
  import { ArrowDown, ArrowUp } from "lucide-svelte";
  import { onMount } from "svelte";
  import { formatMeasureRef } from "./expression-parser";
  import {
    type ExpressionToken,
    filterMeasureMentions,
    IDENTIFIER_CHAR,
    serializeExpressionTokens,
    tokenizeMeasureExpression,
  } from "./measure-mention";

  export let value: string;
  export let id: string;
  // The measures the expression may reference.
  export let measures: MetricsViewSpecMeasure[];
  export let errors: string | undefined = undefined;

  const MENTION = "measure";

  let element: HTMLDivElement;
  let editor: Editor | undefined;
  let focused = false;

  // Set while the "@" suggestion is active.
  let picker: {
    options: MetricsViewSpecMeasure[];
    select: (measure: MetricsViewSpecMeasure) => void;
  } | null = null;
  // Escape closes the picker until the next "@".
  let dismissed = false;
  let focusedIndex = 0;

  // Typing past every match (e.g. "@total_cost + 5") closes the picker too.
  $: open = !!picker && !dismissed && picker.options.length > 0;
  $: options = open ? picker!.options : [];
  $: if (options) focusedIndex = 0;
  // The "@" is not valid in an expression, so the parse error it causes is
  // noise while the picker is open.
  $: shownErrors = open ? undefined : errors;

  function onKeydown(e: KeyboardEvent) {
    if (!open || !picker) return;
    switch (e.key) {
      case "ArrowDown":
        focusedIndex = (focusedIndex + 1) % options.length;
        break;
      case "ArrowUp":
        focusedIndex = (focusedIndex - 1 + options.length) % options.length;
        break;
      case "Enter":
        picker.select(options[focusedIndex]);
        break;
      case "Escape":
        dismissed = true;
        break;
      default:
        return;
    }
    // Keep the key from the editor and from the dialog, which closes on Escape.
    e.preventDefault();
    e.stopPropagation();
  }

  function position(list: HTMLElement) {
    const cleanup = autoUpdate(element, list, () => {
      list.style.width = `${element.getBoundingClientRect().width}px`;
      void computePosition(element, list, {
        strategy: "fixed",
        placement: "bottom-start",
        middleware: [offset(4), flip(), shift({ padding: 8 })],
      }).then(({ x, y }) => {
        Object.assign(list.style, { left: `${x}px`, top: `${y}px` });
      });
    });
    return { destroy: cleanup };
  }

  function docToTokens(doc: ProseMirrorNode): ExpressionToken[] {
    const tokens: ExpressionToken[] = [];
    doc.descendants((node) => {
      if (node.type.name === MENTION) {
        tokens.push({
          type: "measure",
          name: node.attrs.id as string,
          displayName: node.attrs.label as string,
        });
        return false;
      }
      if (node.isText) tokens.push({ type: "text", text: node.text ?? "" });
      return true;
    });
    return tokens;
  }

  onMount(() => {
    editor = new Editor({
      element,
      extensions: [
        // A single paragraph: expressions are one line.
        Document.extend({ content: "paragraph" }),
        Paragraph,
        Text,
        Placeholder.configure({
          placeholder: m.dashboard_pivot_ephemeral_expression_placeholder(),
        }),
        // Enter selects from the picker (handled on the wrapper) and does
        // nothing otherwise.
        Extension.create({
          name: "singleLine",
          addKeyboardShortcuts: () => ({ Enter: () => true }),
        }),
        Mention.extend({ name: MENTION }).configure({
          HTMLAttributes: { class: "measure-chip" },
          deleteTriggerWithBackspace: true,
          renderText: ({ node }) => formatMeasureRef(node.attrs.id as string),
          renderHTML: ({ options, node }) => [
            "span",
            { ...options.HTMLAttributes, title: node.attrs.id as string },
            node.attrs.label as string,
          ],
          suggestion: {
            char: "@",
            allowSpaces: true,
            allowedPrefixes: null,
            allow: ({ state, range }) =>
              !IDENTIFIER_CHAR.test(
                state.doc.textBetween(Math.max(range.from - 1, 0), range.from),
              ),
            items: ({ query }) => filterMeasureMentions(measures, query),
            render: () => {
              const update = (props: {
                items: MetricsViewSpecMeasure[];
                command: (attrs: { id: string; label: string }) => void;
              }) => {
                picker = {
                  options: props.items,
                  select: (measure) =>
                    props.command({
                      id: measure.name ?? "",
                      label: measure.displayName || measure.name || "",
                    }),
                };
              };
              return {
                onStart: (props) => {
                  dismissed = false;
                  update(props);
                },
                onUpdate: update,
                onExit: () => {
                  picker = null;
                },
              };
            },
          },
        }),
      ],
      content: {
        type: "doc",
        content: [
          {
            type: "paragraph",
            content: tokenizeMeasureExpression(value, measures).map((token) =>
              token.type === "text"
                ? { type: "text", text: token.text }
                : {
                    type: MENTION,
                    attrs: { id: token.name, label: token.displayName },
                  },
            ),
          },
        ],
      },
      editorProps: {
        attributes: {
          id,
          role: "textbox",
          "aria-label": m.dashboard_pivot_ephemeral_expression_label(),
          "aria-multiline": "false",
        },
      },
      onUpdate: ({ editor }) => {
        value = serializeExpressionTokens(docToTokens(editor.state.doc));
      },
      onFocus: () => (focused = true),
      onBlur: () => (focused = false),
    });

    return () => editor?.destroy();
  });
</script>

<!-- Keys are intercepted in the capture phase so the picker handles them
     before the editor does. -->
<!-- svelte-ignore a11y-no-static-element-interactions -->
<div class="flex flex-col gap-y-1" on:keydown|capture={onKeydown}>
  <InputLabel
    {id}
    label={m.dashboard_pivot_ephemeral_expression_label()}
    hint={m.dashboard_pivot_ephemeral_expression_hint()}
  />

  <div
    bind:this={element}
    class="editor"
    class:focused
    class:error={!!shownErrors}
  ></div>

  {#if shownErrors}
    <div class="text-red-500 text-xs">{shownErrors}</div>
  {/if}

  {#if open}
    <div
      use:position
      role="listbox"
      aria-label={m.dashboard_pivot_ephemeral_mention_label()}
      class="measure-picker"
    >
      <div class="options">
        {#each options as measure, i (measure.name)}
          <button
            type="button"
            role="option"
            aria-selected={i === focusedIndex}
            class="option"
            class:focused={i === focusedIndex}
            on:mousemove={() => (focusedIndex = i)}
            on:mousedown|preventDefault
            on:click={() => picker?.select(measure)}
          >
            <span class="truncate">{measure.displayName || measure.name}</span>
            {#if measure.displayName && measure.displayName !== measure.name}
              <span class="ml-auto pl-2 text-fg-muted truncate">
                {measure.name}
              </span>
            {/if}
          </button>
        {/each}
      </div>
      <div class="navigation">
        <Kbd.Group>
          <Kbd.Root><ArrowUp size="12px" /></Kbd.Root>
          <Kbd.Root><ArrowDown size="12px" /></Kbd.Root>
          <span>{m.dashboard_pivot_ephemeral_mention_navigate()}</span>
          <!-- i18n-ignore: key name -->
          <Kbd.Root><span>Enter</span></Kbd.Root>
          <span>{m.dashboard_pivot_ephemeral_mention_select()}</span>
        </Kbd.Group>
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  /* Mirrors the text Input's frame. */
  .editor {
    @apply w-full min-h-[30px] px-2 py-1 border rounded-[2px] bg-input;
    @apply text-xs text-fg-primary cursor-text;
  }

  .editor.focused {
    @apply border-primary-500 ring-2 ring-primary-100;
  }

  .editor.error {
    @apply border-red-600 ring-1 ring-transparent;
  }

  .editor :global(.tiptap) {
    @apply outline-none leading-[22px] break-words;
  }

  .editor :global(.tiptap p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    @apply text-fg-muted pointer-events-none absolute;
  }

  .editor :global(.measure-chip) {
    @apply inline-block px-1 rounded-sm border;
    @apply bg-secondary-50 border-secondary-200 text-secondary-800;
  }

  .measure-picker {
    @apply fixed top-0 left-0 z-50 flex flex-col p-1.5;
    @apply border rounded-md bg-popover text-popover-foreground shadow-md;
  }

  .options {
    @apply flex flex-col max-h-48 overflow-y-auto;
  }

  .option {
    @apply flex items-center w-full px-2 py-1.5 rounded-sm text-xs text-left;
  }

  .option.focused {
    @apply bg-popover-accent;
  }

  .navigation {
    @apply flex items-center pt-1.5 px-1.5 text-xs text-popover-foreground/60;
  }
</style>
