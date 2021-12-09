<script>
import { onMount } from "svelte";
import hljs from 'highlight.js';
export let data;
let ref;

onMount(() => {
    hljs.highlightElement(ref);
})

$: console.log(data)

</script>

<pre>
    <code class="language-json" bind:this={ref}>
        {JSON.stringify(data, null, 2)}
    </code>
</pre>

<style>
pre {
    padding:0px;
}
</style>