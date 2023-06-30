The `<ButtonToggleGroup>` components allows the creation of button group. The main `<ButtonToggleGroup>` component must have one or more `<SubButton>` children, and no other direct descendants should be present. Each `<SubButton>` should have a `value: number | string` prop, and the button `<ButtonToggleGroup>` should have a `selected: (number | string)[]` array that determines which buttons are shown as active. The `<ButtonToggleGroup>` may also have a `disabled: (number | string)[]` array.

When a non-disabled subbutton is clicked the `<ButtonToggleGroup>` dispatches the event "click". This will return the `(number | string)` keys of the clicked subbuttons.

This is a fully controlled componentent, so it us up to the containing component to respond to click events dispatched by this component and to pass in a new `selected` prop to update the selection state. If radio button behavior is desired, it is up to the containing component to guarantee exclusive selection when updating `selected`. Likewise, if it is desired that at least one value always be active it's up to the containing component to enforce that constraint.

Additionally, `<SubButton>` may have a `tooltips` prop, with type:

``` javascript
tootips: {
    selected?: string;
    unselected?: string;
    disabled?: string;
  };
```
If the subbutton is disabled, then that will override the other two tooltip options. If  the button is not disabled, the tooltip string corresponding to the buttons selection state will be shown.

# Usage Example:

```javascript
<ButtonToggleGroup selected={[1,2]} disabled={[4]} >
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



