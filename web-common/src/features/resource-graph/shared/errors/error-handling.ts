/**
 * Error handling utilities for resource graph.
 *
 * This module provides:
 * - Typed error classes for different failure scenarios
 * - Error recovery strategies
 * - User-friendly error messages
 * - Error reporting and logging
 */

import { debugLog } from "../config";

/**
 * Base error class for resource graph errors.
 */
export class ResourceGraphError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly recoverable: boolean = true,
  ) {
    super(message);
    this.name = "ResourceGraphError";
  }

  /**
   * Get a user-friendly error message.
   */
  getUserMessage(): string {
    return this.message;
  }

  /**
   * Get suggested recovery action.
   */
  getRecoveryAction(): string | null {
    return this.recoverable ? "Please try again or refresh the page." : null;
  }
}

/**
 * Error thrown when resource data is invalid or missing.
 */
export class ResourceDataError extends ResourceGraphError {
  constructor(
    message: string,
    public readonly resourceId?: string,
  ) {
    super(message, "RESOURCE_DATA_ERROR", true);
    this.name = "ResourceDataError";
  }

  getUserMessage(): string {
    if (this.resourceId) {
      return `Unable to load resource "${this.resourceId}". ${this.message}`;
    }
    return `Unable to load resource data. ${this.message}`;
  }

  getRecoveryAction(): string {
    return "Try refreshing the page or checking if the resource still exists.";
  }
}

/**
 * Error thrown when graph layout fails.
 */
export class GraphLayoutError extends ResourceGraphError {
  constructor(
    message: string,
    public readonly nodeCount: number = 0,
  ) {
    super(message, "GRAPH_LAYOUT_ERROR", true);
    this.name = "GraphLayoutError";
  }

  getUserMessage(): string {
    return `Failed to layout graph with ${this.nodeCount} nodes. ${this.message}`;
  }

  getRecoveryAction(): string {
    return "Try clearing the cache or simplifying the graph.";
  }
}

/**
 * Error thrown when cache operations fail.
 */
export class CacheError extends ResourceGraphError {
  constructor(
    message: string,
    public readonly operation: string,
  ) {
    super(message, "CACHE_ERROR", true);
    this.name = "CacheError";
  }

  getUserMessage(): string {
    return `Cache operation "${this.operation}" failed. ${this.message}`;
  }

  getRecoveryAction(): string {
    return "Try clearing your browser cache or using incognito mode.";
  }
}

/**
 * Error thrown when navigation fails.
 */
export class NavigationError extends ResourceGraphError {
  constructor(
    message: string,
    public readonly targetUrl?: string,
  ) {
    super(message, "NAVIGATION_ERROR", true);
    this.name = "NavigationError";
  }

  getUserMessage(): string {
    return `Failed to navigate to graph. ${this.message}`;
  }

  getRecoveryAction(): string {
    return "Try using the back button and navigating again.";
  }
}

/**
 * Error severity levels.
 */
export enum ErrorSeverity {
  /**
   * Informational - no action needed.
   */
  INFO = "info",

  /**
   * Warning - something unexpected but handled.
   */
  WARNING = "warning",

  /**
   * Error - operation failed but app is still functional.
   */
  ERROR = "error",

  /**
   * Critical - major failure, app may not work correctly.
   */
  CRITICAL = "critical",
}

/**
 * Error context for reporting and debugging.
 */
export interface ErrorContext {
  /**
   * Component or module where error occurred.
   */
  component: string;

  /**
   * Operation that was being performed.
   */
  operation: string;

  /**
   * Additional context data.
   */
  data?: Record<string, any>;

  /**
   * Error severity.
   */
  severity: ErrorSeverity;

  /**
   * Whether to show to user (vs. just log).
   */
  showToUser: boolean;
}

/**
 * Error handler function type.
 */
export type ErrorHandler = (error: Error, context: ErrorContext) => void;

/**
 * Global error handler registry.
 */
const errorHandlers: ErrorHandler[] = [];

/**
 * Register an error handler.
 * Returns a function to unregister.
 */
export function registerErrorHandler(handler: ErrorHandler): () => void {
  errorHandlers.push(handler);
  return () => {
    const index = errorHandlers.indexOf(handler);
    if (index !== -1) {
      errorHandlers.splice(index, 1);
    }
  };
}

/**
 * Report an error through all registered handlers.
 */
export function reportError(error: Error, context: ErrorContext): void {
  // Always log to console
  const logFn =
    context.severity === ErrorSeverity.CRITICAL
      ? console.error
      : context.severity === ErrorSeverity.ERROR
        ? console.error
        : context.severity === ErrorSeverity.WARNING
          ? console.warn
          : console.log;

  logFn(
    `[ResourceGraph:${context.component}] ${context.operation} failed`,
    error,
    context.data,
  );

  debugLog(
    "Error",
    `${context.severity.toUpperCase()} in ${context.component}.${context.operation}`,
    { error, context },
  );

  // Call all registered handlers
  for (const handler of errorHandlers) {
    try {
      handler(error, context);
    } catch (handlerError) {
      console.error("Error handler failed:", handlerError);
    }
  }
}

/**
 * Wrap an operation with error handling.
 */
export function withErrorHandling<T>(
  operation: () => T,
  context: Partial<ErrorContext> & { component: string; operation: string },
): T | null {
  try {
    return operation();
  } catch (error) {
    reportError(error as Error, {
      severity: ErrorSeverity.ERROR,
      showToUser: true,
      ...context,
    });
    return null;
  }
}

/**
 * Wrap an async operation with error handling.
 */
export async function withAsyncErrorHandling<T>(
  operation: () => Promise<T>,
  context: Partial<ErrorContext> & { component: string; operation: string },
): Promise<T | null> {
  try {
    return await operation();
  } catch (error) {
    reportError(error as Error, {
      severity: ErrorSeverity.ERROR,
      showToUser: true,
      ...context,
    });
    return null;
  }
}

/**
 * Create a safe version of a function that won't throw.
 */
export function makeSafe<Args extends any[], Return>(
  fn: (...args: Args) => Return,
  context: Partial<ErrorContext> & { component: string; operation: string },
  fallback: Return,
): (...args: Args) => Return {
  return (...args: Args) => {
    try {
      return fn(...args);
    } catch (error) {
      reportError(error as Error, {
        severity: ErrorSeverity.ERROR,
        showToUser: false,
        ...context,
        data: { args, ...context.data },
      });
      return fallback;
    }
  };
}

/**
 * Get user-friendly error message from any error.
 */
export function getUserErrorMessage(error: unknown): string {
  if (error instanceof ResourceGraphError) {
    return error.getUserMessage();
  }

  if (error instanceof Error) {
    return error.message || "An unexpected error occurred.";
  }

  return "An unexpected error occurred.";
}

/**
 * Get recovery action from any error.
 */
export function getRecoveryAction(error: unknown): string | null {
  if (error instanceof ResourceGraphError) {
    return error.getRecoveryAction();
  }

  return "Please try again or refresh the page.";
}

/**
 * Check if error is recoverable.
 */
export function isRecoverable(error: unknown): boolean {
  if (error instanceof ResourceGraphError) {
    return error.recoverable;
  }

  // Assume unknown errors are recoverable
  return true;
}
