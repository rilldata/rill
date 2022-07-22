<script lang="ts">
  import { flip } from "svelte/animate";
  import Filter from "./Filter.svelte";
  import FilterSet from "./FilterSet.svelte";

  let actives = [
    "Manhattan",
    "Astoria",
    "Brooklyn",
    "Mission",
    "Silver Lake",
    "Extremely Long Neighborhood Name That will tRuncate",
    "Echo Park",
    "Hyde Park",
    "Hockney",
  ];
  let input = "";

  let focused = false;
  function handle(event) {
    if (focused && event.keyCode === 13) {
      if (input.length) {
        actives = [...actives, input];
        const randomIndex = ~~(Math.random() * things.length);
        things[randomIndex].values = [...things[randomIndex].values, input];
        input = "";
      }
    }
  }

  function createGroups(arr, numGroups) {
    const perGroup = Math.ceil(arr.length / numGroups);
    return new Array(numGroups)
      .fill(null)
      .map((_, i) => arr.slice(i * perGroup, (i + 1) * perGroup));
  }

  let filters = [];
  let things = [
    { name: "UKish", values: ["Hockney", "Olympia", "Mission"] },
    {
      name: "NYish",
      values: ["Brooklyn", "williamsburg", "super long neighborhood name"],
    },
    { name: "LAish", values: ["another", "mission", "silver lake"] },
  ];

  let duration = 200;
</script>

<svelte:window on:keypress={handle} />
<input
  class="border  border-gray-500 active:border-blue-500"
  on:focus={() => {
    focused = true;
  }}
  on:blur={() => {
    focused = false;
  }}
  bind:value={input}
/>

<button
  on:click={() => {
    things[1].values = things[1].values.sort((a, b) => Math.random() - 0.5);
  }}>flip</button
>

<div>
  <FilterSet width="900px" style="flex">
    {#each things as { name, values } (name)}
      <div animate:flip={{ duration }}>
        <FilterSet style="inline-flex" width="max-content" direction="row">
          <div>
            {name}
          </div>
          {#each values.filter((si) => !filters.includes(si)) as si (si)}
            <div>
              <Filter
                on:click={() => {
                  filters = [...filters, si];
                }}
                collapseDirection="horizontal">{si}</Filter
              >
            </div>
          {/each}
        </FilterSet>
      </div>
    {/each}
  </FilterSet>
</div>

<div class="grid gap-y-6 pt-6" style:width="1200px">
  flows left to right
  <FilterSet>
    {#each actives as borough, i (borough)}
      <div animate:flip={{ duration: 200 }}>
        <Filter
          collapseDirection="horizontal"
          on:click={() => {
            actives = actives.filter((a) => a !== borough);
          }}>{borough}</Filter
        >
      </div>
    {/each}
  </FilterSet>

  flows top to bottom
  <FilterSet direction="col" height="200px" width="max-content">
    {#each actives as borough, i (borough)}
      <div animate:flip={{ duration: 200 }}>
        <Filter
          collapseDirection="vertical"
          on:click={() => {
            actives = actives.filter((a) => a !== borough);
          }}>{borough}</Filter
        >
      </div>
    {/each}
  </FilterSet>
</div>
