<script>
import { createEventDispatcher } from "svelte";
import { fly, scale, slide } from "svelte/transition";
import { dropStore } from "$lib/drop-store";
import Editor from '$lib/components/Editor.svelte';

export let end = false;
export let padTop = true;

let active = false;
const dispatch = createEventDispatcher();
$: dragging = $dropStore !== undefined;
// bind editorHeight, which is calculated
// in the Editor component with a ResizeObserver.
let editorHeight;
$: dropzoneHeight = !active ? '0px' : `calc(1rem + ${editorHeight}px)`;
$: height = !active ? '1rem' : `calc(1rem + ${editorHeight}px)`;''
</script>

<div style="height: {dropzoneHeight}; transition: height 200ms;">
    {#if !end}
        <div class={dragging && !active ? 'border border-gray transition-all' : ''} style='height:0px; transform: translateY(-4px);'></div>
    {/if}
    <div
        class='
            italic
            {active ? 'pb-2' : ''}
            w-full ease-in {active?'':''}'
        style="
            font-size: 12px;
            height: {height};
            transition: height 150ms;
            transform: translateY({end ? '0' : '-1rem'});"
        on:dragenter|preventDefault={(evt) => {
            active = true;
        }}
        on:dragleave|preventDefault={(evt) => {
            active = false;
        }}
        on:dragover|preventDefault
        on:drop|preventDefault={() => {
            active = false;
            if ($dropStore.type === 'source-to-query') {
                dispatch('source-drop', $dropStore);   
            }
        }}
        >
            {#if active}
            <div class='{padTop ? "pt-3" : ''} pb-3' style='pointer-events: none;'>
                <div transition:slide|local={{duration: 100 }} style="filter: grayscale(100%); opacity: .5;">
                    <Editor bind:editorHeight content={$dropStore.props.content} name="+ new query" />
                </div>
            </div>
            {/if}
    </div>
</div>
