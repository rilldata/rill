<script>
  import {
    Divider,
    Menu,
    MenuItem,
  } from "@rilldata/web-local/lib/components/menu";

  import CheckCircle from "@rilldata/web-common/components/icons/CheckCircle.svelte";
  import CircleEmpty from "@rilldata/web-common/components/icons/EmptyCircle.svelte";

  let commands = [];
  function add(cmd) {
    commands.unshift(cmd);
    commands = commands.slice(0, 10);
  }

  let active = true;
</script>

<button
  on:click={() => {
    active = !active;
  }}
>
  is active? {active}
</button>

{#if active}
  <Menu
    on:escape={() => {
      active = false;
    }}
    on:item-select={() => {
      active = false;
    }}
  >
    <MenuItem on:select={() => add("delete")}>
      <svelte:component
        this={commands[0] === "delete" ? CheckCircle : CircleEmpty}
        slot="icon"
      />
      delete
      <svelte:fragment slot="right">something</svelte:fragment>
    </MenuItem>
    <MenuItem on:select={() => add("save")}>save</MenuItem>
    <Divider />
    <MenuItem on:select={() => add("sort by")}>sort by name</MenuItem>
  </Menu>
{/if}

<ul>
  {#each commands as cmd}
    <li>{cmd}</li>
  {/each}
</ul>

<div>This is a simple mouseover element.</div>
