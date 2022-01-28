<script lang="ts">
import { onMount, createEventDispatcher } from 'svelte';
import {EditorView} from "@codemirror/view";
import {RangeSet} from "@codemirror/rangeset"
import {EditorState, StateField, StateEffect} from "@codemirror/state";
import {gutter, GutterMarker} from "@codemirror/gutter"
import { basicSetup } from "@codemirror/basic-setup";

const dispatch = createEventDispatcher();

export let content;
export let name;
export let editorHeight = 0;

$: editorHeight = componentContainer?.offsetHeight || 0;

let oldContent = content;
let cursorLocation = 0;

let componentContainer;
let editorContainerComponent;
let editor;
let editingTitle = false;
let titleInput;
let titleInputValue;

onMount(() => {
    editor = new EditorView({
        state: EditorState.create({doc: oldContent, extensions: [
            basicSetup,
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

<div class="metrics-editor h-full" bind:this={componentContainer}>
    <div>
        <button on:click={() => dispatch('process')}>Process</button>
        <button on:click={() => dispatch('save')}>Save</button>
        <button on:click={() => dispatch('delete')}>Delete</button>
        <button on:click={() => dispatch('cancel')}>Cancel</button>
    </div>
    <input 
    bind:this={titleInput} 
    on:input={(evt) => {
        titleInputValue = evt.target.value;
        editingTitle = true;
    }} 
    on:blur={()  => { editingTitle = false; }}
    value={name} 
    size={Math.max((editingTitle ? titleInputValue : name)?.length || 0, 5) + 2} 
    on:change={(e) => {dispatch('rename', e.target.value)} } />

    <div class='cm-container bg-white m-5 rounded p-3' bind:this={editorContainerComponent}></div>
    <slot name="prototype-container" />
</div>

<style>
:global(.cm-container .cm-content, .cm-container .cm-gutter) {
    min-height: 400px;
}
.metrics-editor {
    font-size: 12px;
}
</style>