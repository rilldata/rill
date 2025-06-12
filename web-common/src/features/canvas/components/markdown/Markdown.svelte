<script lang="ts">
  import type { MarkdownCanvasComponent } from "./";
  import { getPositionClasses } from "./util";
  import { gfm } from "@milkdown/kit/preset/gfm";
  import {
    Editor,
    rootCtx,
    defaultValueCtx,
    editorViewCtx,
    editorViewOptionsCtx,
    EditorStatus,
  } from "@milkdown/kit/core";
  import { listener, listenerCtx } from "@milkdown/kit/plugin/listener";
  import { commonmark } from "@milkdown/kit/preset/commonmark";
  import { nord } from "@milkdown/theme-nord";
  import { history } from "@milkdown/plugin-history";
  import { replaceAll } from "@milkdown/kit/utils";
  import "./markdown.css";
  import { placeholderCtx, placeholder } from "./placeholder";
  import { onMount } from "svelte";
  import { get } from "svelte/store";

  export let component: MarkdownCanvasComponent;
  export let editable = false;

  let element: HTMLElement;
  let editor: Editor;

  onMount(() => {
    const specStore = get(component.specStore);

    const initialContent = specStore.content;

    editor = Editor.make()
      .config((ctx) => {
        // Mount the editor to the DOM
        ctx.set(rootCtx, element);

        // Set the initial content
        ctx.set(defaultValueCtx, initialContent);

        // Disable/enable readonly mode
        ctx.set(editorViewOptionsCtx, { editable: () => editable });

        // Listen for changes in the editor
        ctx.get(listenerCtx).markdownUpdated((ctx, newMarkdown) => {
          if (ctx.get(editorViewCtx).hasFocus()) {
            component.updateProperty("content", newMarkdown, true);
          }
        });

        ctx.set(placeholderCtx, "Text");
      })
      .config(nord)
      .use(commonmark)
      .use(placeholder)
      .use(gfm)
      .use(history)
      .use(listener);

    editor.create().catch((error) => {
      console.error("Error creating editor:", error);
    });

    return async () => {
      await editor.destroy();
    };
  });

  $: ({ specStore } = component);
  $: markdownProperties = $specStore;

  $: yamlMarkdownContent = markdownProperties.content;

  $: positionClasses = getPositionClasses(markdownProperties.alignment);

  $: if (editor?.status === EditorStatus.Created)
    updateEditor(yamlMarkdownContent);

  function updateEditor(newYamlContent: string) {
    if (editor.ctx.get(editorViewCtx).hasFocus()) {
      return;
    }

    editor.action(replaceAll(newYamlContent));
  }
</script>

<div class="size-full px-2 bg-surface pointer-events-none">
  <div
    role="presentation"
    class="{positionClasses} h-full flex flex-col min-h-min pointer-events-auto"
    class:cursor-text={editable}
    bind:this={element}
    on:click={() => {
      if (editable && editor?.status === EditorStatus.Created) {
        editor.ctx.get(editorViewCtx).focus();
      }
    }}
  ></div>
</div>
