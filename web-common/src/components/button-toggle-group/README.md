The `<ButtonToggleGroup>` components allows the creation of button group. The main `<ButtonToggleGroup>` component must have one or more `<SubButton>` children, and no other direct descendants should be present. Each `<SubButton>` should have a `value: number | string` prop, and the button `<ButtonToggleGroup>` should have a `values: (number | string)[]` array that determines which buttons are shown as active. The `<ButtonToggleGroup>` may also have a `disabled: (number | string)[]` array.

When a non-disabled subbutton is clicked the `<ButtonToggleGroup>` dispatches the event "click". This will return the `(number | string)` keys of the clicked subbuttons.

This is a fully controlled componentent, so it us up to the containing component to respond to click events dispatched by this component and to pass in new `values` to update the selection state. If radio button behavior is desired, it is up to the containing component to guarantee exclusive selection when updating `values`. Likewise, if it is desired that at least one value always be active it's up to the containing component to enforce that.

Additionally, `<SubButton>` may have a `tooltips` prop, with type:

``` javascript
tootips: {
    selected?: string;
    unselected?: string;
    disabled?: string;
  };
```
If the subbutton is disabled, then that will override the other two tooltip options. If however the button is not disabled, the tooltip string corresponding to the buttons selection state will be shown.



# Usage Example:

```javascript
<ButtonToggleGroup values={[1,2]} disabled={[4]} >
  <SubButton value={1}>
    <Bold />%
  </SubButton>
  <SubButton value={2}>
    <Italic />%
  </SubButton>
  <SubButton value={3}>
    <Underline />%
  </SubButton>
  <SubButton value={4}>
    <Strike />%
  </SubButton>
</ButtonToggleGroup>
```

This implementation is based on a pattern recommended by Rich Harris here:
https://stackoverflow.com/questions/56808584/iterate-over-slots-children-in-svelte-3

https://svelte.dev/repl/8e68120858e5322272dc9136c4bb79cc?version=3.5.1

Core API influenced by https://mui.com/material-ui/react-toggle-button/#standalone-toggle-button


---

`<ButtonToggleGroup>` exposes the following props:



`disabledKeys: (number | string)[] = [];` -- An array containing the keys of any sub buttons that are disabled.


`<ButtonToggleGroup>` Dispatches the event "selected-subbutton" when a subbutton is pressed. This will return either the key of the selected button, or a `null` if all subbuttons are deselected as a result of the subbutton press.


It is the responsibility of the containing component to handle both of these events and manage any state external to the component.

---

Each `<SubButton>` must have a unique `valus` prop (`number | string`), Which will be used to determine which sub button is selected, as well as being used as the key that is returned by the events that the `<ButtonToggleGroup>` dispatches.





