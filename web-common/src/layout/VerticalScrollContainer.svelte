<!-- @component
This simple layout component is used to wrap content that should be scrollable, but which is contained in a parent element that has a fixed height and `overflow-hidden`.

This is required as of 2023-05-11 becuase of the way in which the layout of the dashboard component is specified -- it has two columns (measures and leaderboards) that can be of arbitrary height, and in order for them to be independently scrollable, they need to be wrapped in a container that has `overflow: hidden` set so that each of these columns can manage its own scroll behavior.

That in turn means that when other items can occupy the same slot that the dashboard does, those items also need to be responsible for their own scrolling. At this point, that only come up in the cloud views.

This could probably simplified away in the future, but for now, this is the simplest way to get the desired behavior.

https://github.com/rilldata/rill-developer/pull/2398#issuecomment-1544671818
-->

<div class="overflow-auto h-full"><slot /></div>
