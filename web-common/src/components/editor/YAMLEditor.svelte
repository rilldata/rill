<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { onMount } from "svelte";
  import { base } from "./presets/base";
  import { yaml } from "./presets/yaml";
  import { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifacts";

  export let content: string;
  export let extensions: Extension[] = [];
  export let view: EditorView | undefined = undefined;
  export let key: string;
  export let fileArtifact: FileArtifact;
  export let autoSave = true;

  $: ({ saveLocalContent, updateLocalContent } = fileArtifact);

  let container: HTMLElement;

  onMount(() => {
    view = new EditorView({
      state: EditorState.create({
        doc: content,
        extensions: [
          // any extensions passed as props
          ...extensions,
          // establish a basic editor
          base(),
          // establish the yaml editor, which currently only has
          // syntax highlighting
          yaml(),

          EditorView.updateListener.of(({ docChanged, state }) => {
            if (docChanged) {
              const latest = state.doc.toString();
              updateLocalContent(latest);
              if (autoSave) {
                saveLocalContent().catch(console.error);
              }
            }
          }),
        ],
      }),
      parent: container,
    });
  });
</script>

<div bind:this={container} class="contents" />
