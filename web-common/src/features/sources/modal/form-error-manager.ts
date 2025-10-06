import type { FormError } from "./form-state-manager";

export interface ErrorDetails {
  message: string;
  details?: string;
}

export class FormErrorManager {
  private errors: Map<string, FormError> = new Map();

  /**
   * Set an error for a specific form
   */
  setError(formId: string, error: FormError | null): void {
    if (error === null) {
      this.errors.delete(formId);
    } else {
      this.errors.set(formId, error);
    }
  }

  /**
   * Get an error for a specific form
   */
  getError(formId: string): FormError | null {
    return this.errors.get(formId) || null;
  }

  /**
   * Clear an error for a specific form
   */
  clearError(formId: string): void {
    this.errors.delete(formId);
  }

  /**
   * Clear all errors
   */
  clearAllErrors(): void {
    this.errors.clear();
  }

  /**
   * Get all errors as an array
   */
  getAllErrors(): Array<{ formId: string; error: FormError }> {
    return Array.from(this.errors.entries()).map(([formId, error]) => ({
      formId,
      error,
    }));
  }

  /**
   * Check if there are any errors
   */
  hasErrors(): boolean {
    return this.errors.size > 0;
  }

  /**
   * Check if a specific form has an error
   */
  hasError(formId: string): boolean {
    return this.errors.has(formId);
  }

  /**
   * Get error count
   */
  getErrorCount(): number {
    return this.errors.size;
  }

  /**
   * Handle different types of errors and convert them to FormError format
   */
  static parseError(error: unknown): FormError {
    if (error instanceof Error) {
      return {
        message: error.message,
        details: undefined,
      };
    }

    if (error && typeof error === "object" && "message" in error) {
      const errorObj = error as any;
      return {
        message: errorObj.message,
        details:
          errorObj.details !== errorObj.message ? errorObj.details : undefined,
      };
    }

    if (error && typeof error === "object" && "response" in error) {
      const responseError = error as any;
      if (responseError.response?.data) {
        const originalMessage = responseError.response.data.message;
        return {
          message: originalMessage,
          details: undefined,
        };
      }
    }

    return {
      message: "Unknown error occurred",
      details: undefined,
    };
  }

  /**
   * Handle API response errors with human-readable messages
   */
  static parseApiError(
    error: unknown,
    connectorName: string,
    humanReadableErrorMessage: (
      connectorName: string,
      code: string,
      message: string,
    ) => string,
  ): FormError {
    if (error && typeof error === "object" && "response" in error) {
      const responseError = error as any;
      if (responseError.response?.data) {
        const originalMessage = responseError.response.data.message;
        const code = responseError.response.data.code;
        const humanReadable = humanReadableErrorMessage(
          connectorName,
          code,
          originalMessage,
        );

        return {
          message: humanReadable,
          details:
            humanReadable !== originalMessage ? originalMessage : undefined,
        };
      }
    }

    return FormErrorManager.parseError(error);
  }

  /**
   * Set error for a specific form using parsed error
   */
  setParsedError(formId: string, error: unknown): void {
    const parsedError = FormErrorManager.parseError(error);
    this.setError(formId, parsedError);
  }

  /**
   * Set API error for a specific form
   */
  setApiError(
    formId: string,
    error: unknown,
    connectorName: string,
    humanReadableErrorMessage: (
      connectorName: string,
      code: string,
      message: string,
    ) => string,
  ): void {
    const parsedError = FormErrorManager.parseApiError(
      error,
      connectorName,
      humanReadableErrorMessage,
    );
    this.setError(formId, parsedError);
  }
}
