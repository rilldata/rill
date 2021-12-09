<script>
import { onMount, createEventDispatcher } from 'svelte';
import {EditorView} from "@codemirror/view";
import {EditorState} from "@codemirror/state";
import { basicSetup } from "@codemirror/basic-setup";
import { sql } from "@codemirror/lang-sql";

import EditIcon from "$lib/components/EditIcon.svelte"

let container;
let titleInput;
export let content;
export let name;
let oldContent = content;
const dispatch = createEventDispatcher();
let editor;

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
        parent: container
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
            <button class='small-action-button' on:click={() => {
                titleInput.focus();
            }}>
                <EditIcon size={12} />
            </button>
            <input bind:this={titleInput} value={name} on:change={(e) => {dispatch('rename', formatModelName(e.target.value))} } />
        </div>
    </div>
    <div class='editor-container'>
        <div bind:this={container} />
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

.edit-text {
    display: grid;
    grid-template-columns: max-content auto 0px;
    align-items: center;
    justify-content: stretch;
    color: hsl(217,20%, 20%);
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
}

.edit-text:after {
    --size: 40px;
    content: '';
    display: block;
    background-color: green;
    background: linear-gradient(to left, hsl(var(--hue), var(--sat), var(--lgt)) 20%, transparent);
    width: var(--size);
    height: 20px;
    transform: translateX(calc(var(--size) * -1));
    right: 0;
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