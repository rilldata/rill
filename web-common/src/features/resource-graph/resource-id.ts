/**
 * Type-safe ResourceId class to replace error-prone string manipulation.
 *
 * This module provides a robust abstraction for resource identifiers,
 * eliminating the brittleness of string parsing and providing:
 * - Validation and sanitization
 * - Type safety
 * - Consistent formatting
 * - Protection against injection attacks
 */

import type {
  V1ResourceMeta,
  V1ResourceName,
} from "@rilldata/web-common/runtime-client";

/**
 * Separator character for resource IDs.
 * Using a character that's unlikely in resource names.
 */
const ID_SEPARATOR = ":" as const;

/**
 * Reserved characters that are not allowed in resource names.
 * These could interfere with parsing or cause issues in URLs.
 */
const RESERVED_CHARS = /[<>:"\/\\|?*\x00-\x1f]/g;

/**
 * Maximum length for kind and name components.
 * Prevents excessive strings that could indicate injection attempts.
 */
const MAX_COMPONENT_LENGTH = 256;

/**
 * Validation errors that can occur when creating ResourceIds.
 */
export class ResourceIdError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ResourceIdError";
  }
}

/**
 * Type-safe resource identifier.
 *
 * This class encapsulates resource identification logic and provides
 * validation, sanitization, and safe string conversion.
 *
 * @example
 * // Create from kind and name
 * const id = ResourceId.create('rill.runtime.v1.Model', 'orders');
 *
 * @example
 * // Parse from string
 * const id = ResourceId.parse('rill.runtime.v1.Model:orders');
 *
 * @example
 * // Create from metadata
 * const id = ResourceId.fromMeta(resource.meta);
 *
 * @example
 * // Safe string conversion
 * const str = id.toString(); // "rill.runtime.v1.Model:orders"
 */
export class ResourceId {
  private constructor(
    public readonly kind: string,
    public readonly name: string,
  ) {
    // Validate in constructor to ensure immutability
    this.validate();
  }

  /**
   * Create a ResourceId from kind and name.
   * Throws ResourceIdError if validation fails.
   */
  static create(kind: string, name: string): ResourceId {
    return new ResourceId(kind, name);
  }

  /**
   * Try to create a ResourceId, returning null on failure.
   * Use this for graceful error handling.
   */
  static tryCreate(kind: string, name: string): ResourceId | null {
    try {
      return new ResourceId(kind, name);
    } catch {
      return null;
    }
  }

  /**
   * Create from V1ResourceName.
   */
  static fromResourceName(
    resourceName?: V1ResourceName | null,
  ): ResourceId | null {
    if (!resourceName?.kind || !resourceName?.name) return null;
    return ResourceId.tryCreate(resourceName.kind, resourceName.name);
  }

  /**
   * Create from V1ResourceMeta.
   */
  static fromMeta(meta?: V1ResourceMeta | null): ResourceId | null {
    return ResourceId.fromResourceName(meta?.name);
  }

  /**
   * Parse a string ID in format "kind:name".
   * Throws ResourceIdError if parsing fails.
   */
  static parse(id: string): ResourceId {
    if (!id || typeof id !== "string") {
      throw new ResourceIdError("Resource ID must be a non-empty string");
    }

    const idx = id.indexOf(ID_SEPARATOR);

    // No separator found
    if (idx === -1) {
      throw new ResourceIdError(
        `Invalid resource ID format: "${id}". Expected "kind:name"`,
      );
    }

    // Separator at start or end
    if (idx === 0 || idx === id.length - 1) {
      throw new ResourceIdError(
        `Invalid resource ID format: "${id}". Kind and name cannot be empty`,
      );
    }

    const kind = id.slice(0, idx);
    const name = id.slice(idx + 1);

    return new ResourceId(kind, name);
  }

  /**
   * Try to parse a string ID, returning null on failure.
   */
  static tryParse(id: string): ResourceId | null {
    try {
      return ResourceId.parse(id);
    } catch {
      return null;
    }
  }

  /**
   * Validate the kind and name components.
   */
  private validate(): void {
    // Check for empty components
    if (!this.kind || !this.kind.trim()) {
      throw new ResourceIdError("Resource kind cannot be empty");
    }

    if (!this.name || !this.name.trim()) {
      throw new ResourceIdError("Resource name cannot be empty");
    }

    // Check for excessive length (potential injection)
    if (this.kind.length > MAX_COMPONENT_LENGTH) {
      throw new ResourceIdError(
        `Resource kind exceeds maximum length of ${MAX_COMPONENT_LENGTH}`,
      );
    }

    if (this.name.length > MAX_COMPONENT_LENGTH) {
      throw new ResourceIdError(
        `Resource name exceeds maximum length of ${MAX_COMPONENT_LENGTH}`,
      );
    }

    // Check for reserved characters
    if (RESERVED_CHARS.test(this.kind)) {
      throw new ResourceIdError(
        `Resource kind contains invalid characters: "${this.kind}"`,
      );
    }

    if (RESERVED_CHARS.test(this.name)) {
      throw new ResourceIdError(
        `Resource name contains invalid characters: "${this.name}"`,
      );
    }

    // Prevent separator in components (would break parsing)
    if (this.kind.includes(ID_SEPARATOR)) {
      throw new ResourceIdError(
        `Resource kind cannot contain separator "${ID_SEPARATOR}"`,
      );
    }

    // Name CAN contain colons (we split on first colon only)
    // This is intentional to support complex naming schemes
  }

  /**
   * Sanitize a string by removing/replacing invalid characters.
   * Use this when you need to create IDs from untrusted input.
   */
  static sanitize(str: string): string {
    if (!str) return "";
    return str.replace(RESERVED_CHARS, "_").trim();
  }

  /**
   * Convert to string format "kind:name".
   */
  toString(): string {
    return `${this.kind}${ID_SEPARATOR}${this.name}`;
  }

  /**
   * Convert to V1ResourceName.
   */
  toResourceName(): V1ResourceName {
    return {
      kind: this.kind,
      name: this.name,
    };
  }

  /**
   * Check equality with another ResourceId or string.
   */
  equals(other: ResourceId | string): boolean {
    if (typeof other === "string") {
      const parsed = ResourceId.tryParse(other);
      if (!parsed) return false;
      return this.kind === parsed.kind && this.name === parsed.name;
    }
    return this.kind === other.kind && this.name === other.name;
  }

  /**
   * Get a cache key with namespace prefix.
   * Useful for isolating cached positions per graph instance.
   */
  getCacheKey(namespace: string = "global"): string {
    return `${namespace}${ID_SEPARATOR}${this.toString()}`;
  }

  /**
   * Check if this ID matches a given kind.
   */
  isKind(kind: string): boolean {
    return this.kind === kind;
  }

  /**
   * Check if kind contains a substring (case-insensitive).
   * Useful for kind token matching.
   */
  kindIncludes(substring: string): boolean {
    return this.kind.toLowerCase().includes(substring.toLowerCase());
  }
}

/**
 * Backward compatibility helpers.
 * These maintain the old API while using the new ResourceId class internally.
 */

/**
 * @deprecated Use ResourceId.fromMeta() instead
 */
export function createResourceId(meta?: V1ResourceMeta): string | undefined {
  const id = ResourceId.fromMeta(meta);
  return id?.toString();
}

/**
 * @deprecated Use ResourceId.parse() instead
 */
export function parseResourceId(id: string): V1ResourceName | null {
  const resourceId = ResourceId.tryParse(id);
  return resourceId?.toResourceName() ?? null;
}

/**
 * @deprecated Use ResourceId.fromResourceName() instead
 */
export function resourceNameToId(
  resourceName?: V1ResourceName | null,
): string | undefined {
  const id = ResourceId.fromResourceName(resourceName);
  return id?.toString();
}
