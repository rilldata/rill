<script>
import { onMount, createEventDispatcher } from 'svelte';
import {EditorView} from "@codemirror/view";
import {EditorState} from "@codemirror/state";
import { basicSetup } from "@codemirror/basic-setup";
import { sql } from "@codemirror/lang-sql";

import EditIcon from "$lib/components/EditIcon.svelte"

const dispatch = createEventDispatcher();
export let content;
export let name;
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

onMount(() => {
    editor = new EditorView({
        state: EditorState.create({doc: oldContent, extensions: [
            basicSetup,
            sql(),
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