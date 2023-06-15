the `<ButtonToggleGroup>` components allows the creation of radio button style controls that are displayed as a button group. this component maintains a bit of internal state to manage the selection of buttons, and dispatches events when those selections change. The main `<ButtonToggleGroup>` component must have one or more `<GroupButton>` children, and no other elements are allowed. 

Example:

```javascript
<ButtonToggleGroup>
  <GroupButton key={1}>
    <Delta />%
  </GroupButton>
  <GroupButton key={2}>
    <PieChart />%
  </GroupButton>
</ButtonToggleGroup>
```

This implementation is based on a pattern recommended by Rich Harris here:
https://stackoverflow.com/questions/56808584/iterate-over-slots-children-in-svelte-3

https://svelte.dev/repl/8e68120858e5322272dc9136c4bb79cc?version=3.5.1


---

`<ButtonToggleGroup>` exposes the following props:

`selectionRequired: bool = false` -- If this is `true`, then one button will always be selected, like a standard radio button. if `false`, then it is possible to untoggle all sub buttons. In either case, a maximum of 1 sub button may be selected

`defaultKey: number | string` -- The key of the default selection.If undefined, the first sub button will be selected by default.

`disabledKeys: (number | string)[] = [];` -- An array containing the keys of any sub buttons that are disabled.


`<ButtonToggleGroup>` Dispatches the following events:

"select-subbutton" -- When a sub button is selected, this event is dispatched with the key of the sub button

"deselect-subbutton" -- When a sub button is deselected this event is dispatched with the key of the sub button. Note that this event may be fired by itself if no selection is required, indicating that the only selected button has been untoggled. If another sub button has been selected, then this event will be dispatched immediately before the selection event.


It is the responsibility of the containing component to handle both of these events and manage any state external to the component.

---

Each `<GroupButton>` must have a unique `key` prop (`number | string`), Which will be used to determine which sub button is selected, as well as being used as the key that is returned by the events that the `<ButtonToggleGroup>` dispatches.

Additionally, `<GroupButton>` may have a `tooltips` prop, with type:

``` javascript
tootips: {
    selected?: string;
    unselected?: string;
    disabled?: string;
  };
```
If the sub button is disabled, then that will override the other two tooltip options. If however the button is not disabled, the tool tip string corresponding to the buttons selection state will be shown.



