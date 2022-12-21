<script context="module">
  const icons = import.meta.globEager(
    "@rilldata/web-common/components/icons/*.svelte"
  );
  const typeIcons = Object.keys(icons).filter((i) => i.includes("Type"));
  const formIcons = Object.keys(icons).filter(
    (i) =>
      i.includes("Check") ||
      i.includes("Cancel") ||
      i.includes("Empty") ||
      i.includes("Alert") ||
      i.includes("Info")
  );

  const viewIcons = Object.keys(icons).filter(
    (i) =>
      i.includes("Source") ||
      i.includes("Model") ||
      i.includes("Metrics") ||
      i.includes("Explore")
  );
  viewIcons.reverse();

  const others = Object.keys(icons).filter(
    (i) =>
      !typeIcons.includes(i) && !formIcons.includes(i) && !viewIcons.includes(i)
  );
</script>

<h1 class="text-xl mb-4">Iconography</h1>

<!-- <div class="flex flex-row flex-wrap w-screen gap-x-16"> -->
<div
  style="
    column-width: 300px;
  "
>
  {#each [["view icons", viewIcons], ["form icons", formIcons], ["data types", typeIcons], ["others", others]] as [name, iconSet]}
    <div style="break-inside: avoid;">
      <h2 class="mb-1 mt-2">{name}</h2>
      <div
        class="grid gap-x-3 gap-y-2 items-center"
        style:grid-template-columns="max-content max-content max-content"
      >
        {#each iconSet as iconName}
          <div style:width="200px">
            {iconName.split("/").slice(-1)[0]}
          </div>
          <div style:font-size="1.25rem">
            <svelte:component this={icons[iconName].default} />
          </div>
          <div style:font-size=".95rem">
            <svelte:component this={icons[iconName].default} />
          </div>
        {/each}
      </div>
    </div>
  {/each}
</div>
