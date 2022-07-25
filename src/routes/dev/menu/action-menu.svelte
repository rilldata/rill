<script lang="ts">
  import Button from "$lib/components/Button.svelte";
  import notification from "$lib/components/notifications";

  import type {
    Alignment,
    Location,
  } from "$lib/components/floating-element/types";
  import MoreHorizontal from "$lib/components/icons/MoreHorizontal.svelte";
  import SimpleActionMenu from "$lib/components/menu/SimpleActionMenu.svelte";

  function cb(message) {
    return () => notification.send({ message: `selected ${message}` });
  }

  const actionMenuChoices = [
    { main: "first option", right: "first", callback: cb("first") },
    { main: "second option", right: "second", callback: cb("second") },
    { main: "third option", right: "third", callback: cb("third") },
  ];

  let location: Location = "right";
  let alignment: Alignment = "start";
</script>

<select bind:value={location}>
  <option value="left">left</option>
  <option value="right">right</option>
  <option value="top">top</option>
  <option value="bottom">bottom</option>
</select>

<select bind:value={alignment}>
  <option value="start">start</option>
  <option value="middle">middle</option>
  <option value="end">end</option>
</select>

<button> Click here to open menu & focus an element. </button>

<div class="w-full h-screen grid place-center place-content-center">
  <SimpleActionMenu
    dark
    {location}
    {alignment}
    options={actionMenuChoices}
    let:active
    let:toggleMenu
  >
    <Button on:click={toggleMenu} type="primary"
      >See Available Actions <MoreHorizontal size="16px" />
    </Button>
  </SimpleActionMenu>
</div>
