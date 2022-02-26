<script>
import { onMount, createEventDispatcher } from 'svelte';
import { EditorView, keymap } from "@codemirror/view";
import {RangeSet} from "@codemirror/rangeset"
import {indentWithTab} from "@codemirror/commands"
import {EditorState, StateField, StateEffect} from "@codemirror/state";
import {gutter, GutterMarker} from "@codemirror/gutter"
import { basicSetup } from "@codemirror/basic-setup";
import { sql } from "@codemirror/lang-sql";

import RemoveCircleDark from "./icons/RemoveCircleDark.svelte";
import EditIcon from "$lib/components/icons/EditIcon.svelte";
import FreezeIcon from "$lib/components/icons/Freeze.svelte";
import TrashIcon from "$lib/components/icons/Trash.svelte";
import ModelIcon from "$lib/components/icons/Code.svelte";

const dispatch = createEventDispatcher();
export let content;
export let name;
export let editable = true;
export let componentContainer;
export let editorHeight = 0;

$: editorHeight = componentContainer?.offsetHeight || 0;
// export let errorLineNumber;
// export let errorLineMessage;

let oldContent = content;

let editor;
let editorContainer;
let editorContainerComponent;
let titleInput;
let titleInputValue;
let editingTitle = false;


function formatModelName(str) {
    let output = str.trim().replaceAll(' ', '_');
    if (!output.endsWith('.sql')) {
        output += '.sql';
    }
    return output;
}

export function refreshContent(newContent) {
    editor.update({changes: {from: 0, to: editor?.doc?.length || 0, insert: newContent}});
}


function getPos(text, lineNumber) {
    if (lineNumber === 1) return 0;
	return text.slice(0, lineNumber - 1).join(' ').length + 1;
}

const breakpointGutter = [
  ///breakpointState,
  gutter({
    class: "cm-breakpoint-gutter",
    // markers: v => v.state.field(breakpointState),
    initialSpacer: () => breakpointMarker,
    domEventHandlers: {
      mousedown(view, line) {
        return true
      }
    }
  }),
  EditorView.baseTheme({
    ".cm-breakpoint-gutter .cm-gutterElement": {
      color: "red",
      paddingLeft: "5px",
      cursor: "default"
    }
  })
]

const breakpointMarker = new class extends GutterMarker {
  toDOM() { 
      const element = document.createElement('div');
      element.className='gutter-indicator'
      const marker = new RemoveCircleDark({
          target: element,
          props: { size: 13 }
      })
      return element;
    }
}

let cursorLocation = 0;

onMount(() => {
    editor = new EditorView({
        state: EditorState.create({doc: oldContent, extensions: [
            basicSetup,
            sql(),
            keymap.of([indentWithTab]),
            breakpointGutter,
            EditorView.updateListener.of((v)=> {
                const candidateLocation = v.state.selection.ranges[0].head;
                if (candidateLocation !== cursorLocation) {
                    cursorLocation = candidateLocation;
                    dispatch('cursor-location', {location: cursorLocation, content: v.state.doc.toString()})
                }
                if (v.focusChanged) {
                    if (v.view.hasFocus) {
                        dispatch('receive-focus');
                    } else {
                        dispatch('release-focus');
                    }
                }
                if(v.docChanged) {
                    dispatch('write', {
                        content: v.state.doc.toString()
                    });
                }
            })
        ]}),
        parent: editorContainerComponent
    });
    const obs = new ResizeObserver(() => {
        editorHeight = componentContainer.offsetHeight;
    })
    obs.observe(componentContainer);
})

</script>

<div bind:this={componentContainer}>
    <div class='editor-container border h-full' bind:this={editorContainer}>
        <div bind:this={editorContainerComponent} />
    </div>
</div>

<style>
.editor-container {
    padding: .5rem;
    background-color: white;
    border-radius: .25rem;
    min-height: 400px;
}

</style>