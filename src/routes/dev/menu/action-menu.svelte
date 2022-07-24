<script lang="ts">
  import Button from "$lib/components/Button.svelte";

  import type {
    Alignment,
    Location,
  } from "$lib/components/floating-element/types";
  import CaretDownIcon from "$lib/components/icons/CaretDownIcon.svelte";
  import SimpleActionMenu from "$lib/components/menu/SimpleActionMenu.svelte";
  import notification from "$lib/components/notifications";
  const actionMenuChoices = [
    { label: "first option", right: "first", value: 5 },
    { label: "second option", right: "second", value: 20 },
    { label: "third option", right: "third", value: 30 },
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
    on:select={(event) => {
      notification.send({ message: `selected ${event.detail.label}` });
    }}
    {location}
    {alignment}
    actions={actionMenuChoices}
    let:active
    let:toggleMenu
  >
    <Button on:click={toggleMenu} type="primary"
      >Click to Expand {active}
      <span class={active ? "-rotate-180" : ""}>
        <CaretDownIcon size="16px" />
      </span>
    </Button>
  </SimpleActionMenu>
</div>
