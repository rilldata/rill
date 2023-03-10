# `data-graphic`

This directory contains components for building flexible data graphics. A solid component system should be:

- _composable_ - the system of domain and range cascading should ensure that we can do lots of complex things easily, such as nesting graphic contexts within each other.
- _mostly-declarative_ - by taking care of a great deal of complexity for the user – and exposing common primitives like scales in the right places – someone shouldn't have to write too much code to make an interactive data graphic.
- _reactive_ - the domain-space (e.g. the scale domains) should be able to reactively respond to changes in the data. The range-space should be able to respond to changes in configuration parameters. And this should be efficient enough that all values should be tweenable!
- _elegant_ - the system should have sensible defaults and high-quality, human design.
- _cohesive_ - ideally, all our charts and plots should fit nicely within Rill's larger design system. One of the benefits of building our own system is that we can achieve this thread without paying for loss of functionality.

This component set is organized as such:

- `elements` - contains the main containers for data graphics.
- `actions` - contains various actions & action factories used for data graphics.
- `constants` - contains the constants used throughout the component set.
- `guides` - guides are components that orient the data graphic, such as axes, grids, and mouseover labels.
- `marks` - contains the main components used to map data to geometric shapes.
- `functional-components` - components that perform some small function and then expose the output in a slot. These convenience components enable users to add a bit of custom functionality when needed without having to resort to reaching into the `script` tag.
- `state` - contains various store factories used throughout the component set.
