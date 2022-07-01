<script context="module">
  const icons = import.meta.globEager("/src/lib/components/icons/*.svelte");
  const typeIcons = Object.keys(icons).filter((i) => i.includes("Type"));
  const formIcons = Object.keys(icons).filter(
    (i) =>
      i.includes("Check") ||
      i.includes("Cancel") ||
      i.includes("Empty") ||
      i.includes("Alert")
  );

  const viewIcons = Object.keys(icons).filter(
    (i) =>
      i.includes("Code") ||
      i.includes("List") ||
      i.includes("Explore") ||
      i.includes("PlayCircle")
  );

  const others = Object.keys(icons).filter(
    (i) =>
      !typeIcons.includes(i) && !formIcons.includes(i) && !viewIcons.includes(i)
  );
</script>

<h1 class="text-xl mb-4">Iconography</h1>

<div>
  {#each [["view icons", viewIcons], ["form icons", formIcons], ["data types", typeIcons], ["others", others]] as [name, iconSet]}
    <h2 class="mb-1 mt-3">{name}</h2>
    <div
      class="grid gap-x-3 gap-y-2 items-center"
      style:grid-template-columns="max-content max-content max-content
      max-content"
    >
      {#each iconSet as iconName}
        <div style:width="200px">
          {iconName.split("/").slice(-1)[0]}
        </div>
        <div style:font-size="2rem">
          <svelte:component this={icons[iconName].default} />
        </div>
        <div style:font-size="1rem">
          <svelte:component this={icons[iconName].default} />
        </div>
        <div style:font-size="12px">
          <svelte:component this={icons[iconName].default} />
        </div>
      {/each}
    </div>
  {/each}
</div>
<!-- <div
  class="grid gap-x-3 gap-y-2 items-center"
  style:grid-template-columns="max-content max-content max-content max-content"
>
  {#each Object.keys(icons) as iconName}
    <div>
      {iconName.split("/").slice(-1)[0]}
    </div>
    <div style:font-size="2rem">
      <svelte:component this={icons[iconName].default} />
    </div>
    <div style:font-size="1rem">
      <svelte:component this={icons[iconName].default} />
    </div>
    <div style:font-size="12px">
      <svelte:component this={icons[iconName].default} />
    </div>
  {/each}
</div> -->
