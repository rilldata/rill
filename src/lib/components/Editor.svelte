<script>
import { onMount, createEventDispatcher } from 'svelte';
import {EditorView} from "@codemirror/view";
import {RangeSet} from "@codemirror/rangeset"
import {EditorState, StateField, StateEffect} from "@codemirror/state";
import {gutter, GutterMarker} from "@codemirror/gutter"
import { basicSetup } from "@codemirror/basic-setup";
import { sql } from "@codemirror/lang-sql";

import RemoveCircleDark from "./RemoveCircleDark.svelte";


import EditIcon from "$lib/components/EditIcon.svelte"

const dispatch = createEventDispatcher();
export let content;
export let name;

export let errorLineNumber;
export let errorLineMessage;

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

const breakpointEffect = StateEffect.define({
  map: (val, mapping) => ({pos: mapping.mapPos(val.pos), on: val.on})
})

const breakpointState = StateField.define({
  create() { return RangeSet.empty },
  update(set, transaction) {
    set = set.map(transaction.changes)
    for (let e of transaction.effects) {
      if (e.is(breakpointEffect)) {
        if (e.value.on)
          set = set.update({add: [breakpointMarker.range(e.value.pos)]})
        else
          set = set.update({filter: from => from != e.value.pos})
      }
    }
    return set
  }
})

function toggleBreakpoint(view, pos) {
  let breakpoints = view.state.field(breakpointState)
  let hasBreakpoint = false
  breakpoints.between(pos, pos, () => {hasBreakpoint = true})
  view.dispatch({
    effects: breakpointEffect.of({pos, on: !hasBreakpoint})
  })
}

const breakpointGutter = [
  breakpointState,
  gutter({
    class: "cm-breakpoint-gutter",
    markers: v => v.state.field(breakpointState),
    initialSpacer: () => breakpointMarker,
    domEventHandlers: {
      mousedown(view, line) {
        console.log(line)
        toggleBreakpoint(view, line.from);
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

let prevError = undefined;
$: if (editor && errorLineNumber) {
    toggleBreakpoint(editor, getPos(editor.state.doc.text, errorLineNumber));
    prevError = errorLineNumber;
} else if (editor && !errorLineNumber && prevError) {
    toggleBreakpoint(editor, getPos(editor.state.doc.text, prevError));
    prevError = undefined;
}

const breakpointMarker = new class extends GutterMarker {
  toDOM() { 
      const element = document.createElement('div');
      element.className='gutter-indicator'
      const marker = new RemoveCircleDark({
          target: element,
          props: { size: 13 }
      })
    //   element.textContent = '!!';
      return element;
    // const element = document.createElement('div');
    // element.className = 'gutter-indicator';
    // element.textContent = "⚠️";
    // return element;
    }
}

onMount(() => {
    editor = new EditorView({
        state: EditorState.create({doc: oldContent, extensions: [
            basicSetup,
            sql(),
            breakpointGutter,
            EditorView.updateListener.of((v)=> {
                if (v.focusChanged) {
                    if (v.view.hasFocus) {
                        dispatch('receive-focus');
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
})

</script>
<div>
    <div class=controls>
        <div class="close-container">
            <button class="small-action-button really-small-round" on:click={() => dispatch('delete')}>✕</button>
        </div>
        <button class=small-action-button on:click={() => dispatch('up')}>↑</button>
        <button class=small-action-button on:click={() => dispatch('down')}>↓</button>
        <div class='edit-text'>
            <input 
                bind:this={titleInput} 
                on:input={(evt) => {
                    titleInputValue = evt.target.value;
                    editingTitle = true;
                }} 
                on:blur={()  => { editingTitle = false; }}
                value={name} 
                size={Math.max((editingTitle ? titleInputValue : name)?.length || 0, 5) + 2} 
                on:change={(e) => {dispatch('rename', formatModelName(e.target.value))} } />
                <button class='small-action-button edit-button' on:click={() => {
                    titleInput.focus();
                }}>
                    <EditIcon size={12} />
                </button>
        </div>
    </div>
    <div class='editor-container' bind:this={editorContainer}>
        <div bind:this={editorContainerComponent} />
    </div>
</div>

<style>
.editor-container {
    padding: .5rem;
    background-color: white;
    border-radius: .25rem;
    box-shadow: 0px .25rem .25rem rgba(0,0,0,.05);
}

.controls {
    display: grid;
    grid-template-columns: max-content max-content max-content auto;
    align-items: stretch;
    align-content: stretch;
    justify-content: stretch;
    justify-items: stretch;
    width: 100%;
    margin-bottom: .25rem;
}

.edit-button {
    opacity: 0;
    transition: opacity 150ms;
}

.edit-text:hover .edit-button {
    opacity: 1;
}

.edit-text {
    max-width: 100%;
    display: grid;
    grid-template-columns: auto max-content;
    align-items: center;
    justify-content: stretch;
    color: hsl(217,20%, 20%);
    padding-left: .5rem;
}


.edit-text input {
    width: 100%;
    font-family: "MD IO 0.4";
    font-size: 12px;
    background-color: transparent;
    border: none;
    color: hsl(217,20%, 50%);
    text-overflow: ellipsis;
    padding: 0;
    box-sizing: border-box;
}

.edit-text input:focus {
    color: hsl(217,20%, 20%);
    outline: none;
}

.close-container {
    padding-left: .25rem;
    padding-right: .25rem;
    display: grid;
    place-items: center;
}

.really-small-round {
    background-color: hsl(217,20%, 95%);
    color: hsl(217,20%, 20%);
    font-size: 10px;
    padding: 0px;
    width: 16px;
    height: 16px;
    border-radius: 50%;
}

.really-small-round:hover {
    background: hsla(var(--hue), var(--sat), 20%, 1);
    color: white;
}
</style>