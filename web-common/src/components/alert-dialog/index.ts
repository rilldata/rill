import { AlertDialog as AlertDialogPrimitive } from "bits-ui";

import Title from "web-common/src/components/alert-dialog/alert-dialog-title.svelte";
import Action from "web-common/src/components/alert-dialog/alert-dialog-action.svelte";
import Cancel from "web-common/src/components/alert-dialog/alert-dialog-cancel.svelte";
import Portal from "web-common/src/components/alert-dialog/alert-dialog-portal.svelte";
import Footer from "web-common/src/components/alert-dialog/alert-dialog-footer.svelte";
import Header from "web-common/src/components/alert-dialog/alert-dialog-header.svelte";
import Overlay from "web-common/src/components/alert-dialog/alert-dialog-overlay.svelte";
import Content from "web-common/src/components/alert-dialog/alert-dialog-content.svelte";
import Description from "web-common/src/components/alert-dialog/alert-dialog-description.svelte";

const Root = AlertDialogPrimitive.Root;
const Trigger = AlertDialogPrimitive.Trigger;

export {
  Root,
  Title,
  Action,
  Cancel,
  Portal,
  Footer,
  Header,
  Trigger,
  Overlay,
  Content,
  Description,
  //
  Root as AlertDialog,
  Title as AlertDialogTitle,
  Action as AlertDialogAction,
  Cancel as AlertDialogCancel,
  Portal as AlertDialogPortal,
  Footer as AlertDialogFooter,
  Header as AlertDialogHeader,
  Trigger as AlertDialogTrigger,
  Overlay as AlertDialogOverlay,
  Content as AlertDialogContent,
  Description as AlertDialogDescription,
};
