<script lang="ts">
  import type { EmbedOptions, Result } from "vega-embed";
  import type { SignalListeners, View, VisualizationSpec } from "./types";

  import { createEventDispatcher, onDestroy } from "svelte";
  import vegaEmbed from "vega-embed";

  import { WIDTH_HEIGHT } from "./constants";
  import {
    addSignalListenersToView,
    combineSpecWithDimension,
    computeSpecChanges,
    removeSignalListenersFromView,
    shallowEqual,
    updateMultipleDatasetsInView,
  } from "./utils";

  export let options: EmbedOptions;
  export let spec: VisualizationSpec;
  export let view: View | undefined;
  export let signalListeners: SignalListeners = {};
  export let data: Record<string, unknown> = {};

  const dispatch = createEventDispatcher();

  let result: Result | undefined = undefined;
  let prevOptions: EmbedOptions = {};
  let prevSignalListeners: SignalListeners = {};
  let prevSpec: VisualizationSpec = {};
  let prevData: Record<string, unknown> = {};
  let chartContainer: HTMLElement;

  $: {
    if (!shallowEqual(data, prevData)) {
      update();
    }
    prevData = data;
  }

  $: {
    if (chartContainer !== undefined) {
      // only create a new view if neccessary
      if (!shallowEqual(options, prevOptions, WIDTH_HEIGHT)) {
        createView();
      } else {
        const specChanges = computeSpecChanges(
          combineSpecWithDimension(spec, options),
          combineSpecWithDimension(prevSpec, prevOptions),
        );
        const newSignalListeners = signalListeners;
        const oldSignalListeners = prevSignalListeners;

        if (specChanges) {
          if (specChanges.isExpensive) {
            createView();
          } else if (result !== undefined) {
            const areSignalListenersChanged = !shallowEqual(
              newSignalListeners,
              oldSignalListeners,
            );
            view = result.view;
            if (specChanges.width !== false) {
              view.width(specChanges.width);
            }
            if (specChanges.height !== false) {
              view.height(specChanges.height);
            }
            if (areSignalListenersChanged) {
              if (oldSignalListeners) {
                removeSignalListenersFromView(view, oldSignalListeners);
              }
              if (newSignalListeners) {
                addSignalListenersToView(view, newSignalListeners);
              }
            }
            view.runAsync();
          }
        } else if (
          !shallowEqual(newSignalListeners, oldSignalListeners) &&
          result !== undefined
        ) {
          view = result.view;
          if (oldSignalListeners) {
            removeSignalListenersFromView(view, oldSignalListeners);
          }
          if (newSignalListeners) {
            addSignalListenersToView(view, newSignalListeners);
          }
          view.runAsync();
        }
      }
      prevOptions = options;
      prevSignalListeners = signalListeners;
      prevSpec = spec;
    }
  }

  onDestroy(() => {
    clearView();
  });

  async function createView() {
    clearView();
    try {
      result = await vegaEmbed(chartContainer, spec, options);
      view = result.view;
      if (addSignalListenersToView(view, signalListeners)) {
        view.runAsync();
      }
      onNewView(view);
    } catch (e) {
      handleError(e as Error);
    }
  }

  function clearView() {
    if (result) {
      result.finalize();
      result = undefined;
      view = undefined;
    }
  }

  function handleError(error: Error) {
    dispatch("onError", {
      error: error,
    });
    console.warn(error);
  }

  function onNewView(view: View) {
    update();
    dispatch("onNewView", {
      view: view,
    });
  }

  async function update() {
    if (data && Object.keys(data).length > 0 && result !== undefined) {
      view = result.view;
      updateMultipleDatasetsInView(view, data);
      await view.resize().runAsync();
    }
  }
</script>

<div bind:this={chartContainer} />
