/**
 * CSS Sanitizer for Theme System
 *
 * Sanitizes CSS variable values from theme definitions to prevent XSS attacks.
 * The backend controls which CSS variable names are allowed, so the frontend
 * only needs to validate the values for safety.
 */

/**
 * Sanitizes CSS variable values from the new theme.light.variables and theme.dark.variables structure
 * @param variables - Object mapping CSS variable names to values
 * @returns Sanitized object with only safe variable values
 */
export function sanitizeThemeVariables(
  variables: Record<string, string> | undefined,
): Record<string, string> {
  if (!variables) return {};

  const sanitized: Record<string, string> = {};

  for (const [name, value] of Object.entries(variables)) {
    // Ensure variable name starts with -- (normalize if needed)
    const normalizedName = name.startsWith("--") ? name : `--${name}`;

    // Validate that value is safe (no dangerous patterns)
    if (!isSafeVariableValue(value)) {
      console.warn(
        `Skipping potentially unsafe CSS variable value for ${normalizedName}: ${value}`,
      );
      continue;
    }

    sanitized[normalizedName] = value;
  }

  return sanitized;
}

/**
 * Converts sanitized theme variables to CSS string
 * @param lightVariables - Light mode variables
 * @param darkVariables - Dark mode variables
 * @param scopeSelector - Optional CSS selector to scope the variables to (e.g., ".dashboard-theme-boundary")
 * @returns CSS string with scoped variables
 */
export function themeVariablesToCSS(
  lightVariables: Record<string, string>,
  darkVariables: Record<string, string>,
  scopeSelector?: string,
): string | null {
  const hasLight = Object.keys(lightVariables).length > 0;
  const hasDark = Object.keys(darkVariables).length > 0;

  if (!hasLight && !hasDark) return null;

  let css = "";

  if (hasLight) {
    const lightSelector = scopeSelector
      ? `${scopeSelector}:not(.dark), :not(.dark) ${scopeSelector}`
      : ":root:not(.dark)";
    css += `${lightSelector} {\n`;
    for (const [name, value] of Object.entries(lightVariables)) {
      css += `  ${name}: ${value};\n`;
    }
    css += "}\n\n";
  }

  if (hasDark) {
    const darkSelector = scopeSelector
      ? `${scopeSelector}.dark, .dark ${scopeSelector}`
      : ":root.dark";
    css += `${darkSelector} {\n`;
    for (const [name, value] of Object.entries(darkVariables)) {
      css += `  ${name}: ${value};\n`;
    }
    css += "}\n";
  }

  return css.trim() || null;
}

/**
 * Validates that a CSS variable value is safe
 * Rejects values containing dangerous patterns that could lead to XSS
 */
function isSafeVariableValue(value: string): boolean {
  const lowerValue = value.toLowerCase();

  // Block dangerous URL schemes
  const dangerousSchemes = [
    "javascript:",
    "data:",
    "vbscript:",
    "file:",
    "about:",
  ];

  for (const scheme of dangerousSchemes) {
    if (lowerValue.includes(scheme)) {
      return false;
    }
  }

  // Block expression() and behavior() (old IE CSS expressions)
  if (lowerValue.includes("expression(") || lowerValue.includes("behavior(")) {
    return false;
  }

  // Block @import statements
  if (lowerValue.includes("@import")) {
    return false;
  }

  // Block external URLs in url()
  if (lowerValue.includes("url(")) {
    // Allow only relative URLs, data URIs for fonts/images are blocked above
    const urlPattern = /url\s*\(\s*['"]?([^'")]+)['"]?\s*\)/gi;
    const urls = [...value.matchAll(urlPattern)];

    for (const urlMatch of urls) {
      const url = urlMatch[1].trim();
      // Block absolute URLs (http://, https://, //)
      if (url.match(/^(https?:)?\/\//)) {
        return false;
      }
    }
  }

  // Block HTML/XML content
  if (lowerValue.includes("<") || lowerValue.includes(">")) {
    return false;
  }

  return true;
}
