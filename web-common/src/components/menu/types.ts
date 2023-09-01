export type SelectMenuItem = {
  // the key of that identifies the item in call
  key: string | number;
  // the main text
  main: string;
  // the secondary text below the main text
  description?: string;
  // text to display to the right of "main"
  right?: string;
  // will this item be used as a divider?
  divider?: boolean;
  // will this item be disabled?
  disabled?: boolean;
};
