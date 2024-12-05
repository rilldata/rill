import type { CanvasEntity } from "@rilldata/web-common/features/canvas/stores/canvas-entity";

/**
 * A CanvasMutatorCallback is a function that mutates
 * a CanvasEntity, i.e., the data single Canvas.
 * This will often be a closure over other parameters
 * that are relevant to the mutation.
 */
export type CanvasMutatorCallback = (canvasEntity: CanvasEntity) => void;

/**
 * CanvasCallbackExecutor is a function that takes a
 * CanvasMutatorCallback and executes it. The
 * CanvasCallbackExecutor is a closure containing a reference
 * to the live Canvas, and therefore calling this function
 * on a CanvasMutatorCallback will actually update the Canvas.
 */
export type CanvasCallbackExecutor = (callback: CanvasMutatorCallback) => void;
