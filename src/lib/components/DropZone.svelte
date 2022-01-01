<script>
import { createEventDispatcher } from "svelte";
import { fly, scale } from "svelte/transition";
import { dropStore } from "$lib/drop-store";
import Editor from '$lib/components/Editor.svelte';

export let end = false;
export let padTop = true;

let active = false;
const dispatch = createEventDispatcher();
$: dragging = $dropStore !== undefined;
$: dropzoneHeight = !active ? ('0px') : ('5rem');
$: height = !active? (!end? '1rem': '24rem') : (!end? '5rem' : '24rem')
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
            transition: height 150ms;
            height: calc({height}); transform: translateY({end ? '0' : '-1rem'});"
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
            <div class='{padTop ? "pt-3" : ''}' style='pointer-events: none;'>
                <div in:fly={{duration: 200, y: -10}} style="filter: grayscale(100%); opacity: .5;">
                    <Editor content={$dropStore.props.content} name="+ new query" />
                </div>
            </div>
            {/if}
    </div>
</div>
