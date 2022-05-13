<script context="module">
  const modules = import.meta.globEager("/src/stories/*.svelte");
  const componentNames = Object.keys(modules);
</script>

<script lang="ts">
  // NOTE: this component was borrowed from a gist
  import { browser } from "$app/env";
  import { afterUpdate, onMount } from "svelte";

  let current = componentNames[0];
  let Component;
  let props = {};
  let viewerEl: HTMLIFrameElement;
  let viewerWidth = null;
  let allowCustomWidth = false;
  let viewerWidthPreset = [360, 480, 720, 800, 1200];
  $: loadComponent(current);

  const loadComponent = async (current) => {
    try {
      // const res = await modules[current]()
      Component = modules[current].default;
      props = modules[current].defaultProps || {};
    } catch (err) {
      console.error(err);
    }
  };
  const setCurrentByName = (name) => () => {
    current = name;
  };

  const onPropsChange = (e) => {
    const value = e.target.propContent.value;
    props = JSON.parse(value);
  };

  const setViewerWidth = (w: number) => () => {
    viewerWidth = w;
  };

  const setAllowCustomWidth = () => {
    allowCustomWidth = !allowCustomWidth;
  };

  let renderedComponent;

  onMount(() => {
    const styles = Array.from(document.head.querySelectorAll("style")).map(
      (node) => node.cloneNode(true)
    );
    viewerEl.contentWindow.document.head.append(...styles);
  });

  afterUpdate(() => {
    if (!Component) return;
    renderedComponent?.$destroy?.();
    renderedComponent = new Component({
      target: viewerEl.contentWindow.document.body,
      props,
    });
  });
</script>

<div data-grid="container" class="h-screen w-full bg-slate-800 text-white">
  <div data-grid="a" class="p-4">
    <h1 class="mb-4 text-xl">Vignettes</h1>
    <ul>
      {#each componentNames as name (name)}
        <li class:opacity-40={name !== current}>
          <button on:click={setCurrentByName(name)}>
            {name.split("/src/stories/")[1] || ""}
          </button>
        </li>
      {/each}
    </ul>
  </div>
  <div data-grid="b" class="p-4">
    <!-- <h1 class="mb-4 text-xl">Props</h1>
        {#if browser && props}
            <form on:submit|preventDefault={onPropsChange}>
                <textarea
                    name="propContent"
                    rows={10}
                    class="w-full p-2 border border-blue-200 font-mono"
                    >{JSON.stringify(props, null, 2)}</textarea
                >
                <button type="submit" class="bg-blue-500 text-white px-4 py-2">Update</button>
            </form>
        {:else}
            No defaultProps found
        {/if} -->
  </div>
  <div
    data-grid="c"
    class="mx-8 mt-8 self-stretch flex-1 flex flex-col items-center overflow-hidden
               bg-gray-100 border-1 border-gray-300 rounded-tl-4xl rounded-tr-4xl rounded p-5"
  >
    <div
      class="h-12 top-0 sticky bg-gray-100 z-50 w-full flex items-center justify-center text-gray-600"
    >
      <menu class="inline-flex">
        <button class="px-2 py-1" on:click={setViewerWidth(null)}>Full</button>
        {#each viewerWidthPreset as w (w)}
          <button
            on:click={setViewerWidth(w)}
            class:bg-blue-500={w === viewerWidth}
            class="px-2 py-1">{w}</button
          >
        {/each}
        <button class="px-2 py-1" on:click={setAllowCustomWidth}>Custom</button>
        {#if allowCustomWidth}
          <input type="number" bind:value={viewerWidth} />
        {/if}
      </menu>
    </div>
    <iframe
      title="viewer"
      class="border-1 border-gray-300 h-full rounded-tl rounded-tr"
      bind:this={viewerEl}
      style={`width: ${viewerWidth ? viewerWidth + "px" : `100%`};`}
    />
  </div>
</div>

<style>
  div[data-grid="container"] {
    display: grid;
    grid-template-columns: 1fr 4fr;
    grid-template-rows: 1fr 1fr;
    grid-template-areas:
      "a c"
      "b c";
  }

  div[data-grid="a"] {
    grid-area: a;
  }

  div[data-grid="b"] {
    grid-area: b;
  }

  div[data-grid="c"] {
    grid-area: c;
  }
</style>
