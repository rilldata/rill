# Menu components

This directory contains our menu components. There are three types of components here:

- `core` – Basic container and logical components for usage. These items do _not_ float on their own; we are mostly using them with `WithTogglableFloatingElement.svelte` component, defined in `src/lib/components/floating-element`.
  - `Menu.svelte` – the parent element. This basic component handles keyboard interactions.
  - `MenuItem.svelte` – the main child element. Handles selection and rendering of slots.
- `wrappers` – The `With<thing>.svelte` collection. Lke a tooltip, these components wrap a DOM element and provide a floating menu component in some way.
  - `WithSelectMenu.svelte` – this utilizes `WithFloatingMenu.svelte` and makes it so that you provide the trigger, but the menu attached to the trigger functions like a selector menu.
- `triggers` – Trigger components used in compositions
  - `SelectButton.svelte` – a composed trigger button for use in `SimpleSelectMenu.svelte`.
- `compositions` – Convenience components comprised of all of the above, ready-made for usage.
  - `SelectMenu.svelte` – this is a full component, combining `SelectButton` and `WithSelectMenu` to create a simple drop-in replacement for `select`. It has other features as well; see the props.