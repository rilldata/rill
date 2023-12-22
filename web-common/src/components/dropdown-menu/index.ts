import { DropdownMenu as DropdownMenuPrimitive } from "bits-ui";
import Item from "./DropdownMenuItem.svelte";
import Label from "./DropdownMenuLabel.svelte";
import Content from "./DropdownMenuContent.svelte";
import Shortcut from "./DropdownMenuShortcut.svelte";
import RadioItem from "./DropdownMenuRadioItem.svelte";
import Separator from "./DropdownMenuSeparator.svelte";
import RadioGroup from "./DropdownMenuRadioGroup.svelte";
import SubContent from "./DropdownMenuSubContent.svelte";
import SubTrigger from "./DropdownMenuSubTrigger.svelte";
import CheckboxItem from "./DropdownMenuCheckboxItem.svelte";

const Sub = DropdownMenuPrimitive.Sub;
const Root = DropdownMenuPrimitive.Root;
const Trigger = DropdownMenuPrimitive.Trigger;
const Group = DropdownMenuPrimitive.Group;

export {
  Sub,
  Root,
  Item,
  Label,
  Group,
  Trigger,
  Content,
  Shortcut,
  Separator,
  RadioItem,
  SubContent,
  SubTrigger,
  RadioGroup,
  CheckboxItem,
  //
  Root as DropdownMenu,
  Sub as DropdownMenuSub,
  Item as DropdownMenuItem,
  Label as DropdownMenuLabel,
  Group as DropdownMenuGroup,
  Content as DropdownMenuContent,
  Trigger as DropdownMenuTrigger,
  Shortcut as DropdownMenuShortcut,
  RadioItem as DropdownMenuRadioItem,
  Separator as DropdownMenuSeparator,
  RadioGroup as DropdownMenuRadioGroup,
  SubContent as DropdownMenuSubContent,
  SubTrigger as DropdownMenuSubTrigger,
  CheckboxItem as DropdownMenuCheckboxItem,
};
