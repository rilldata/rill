/**
 * CSS Sanitizer for Theme System
 * 
 * Extracts only known safe CSS variables from user-provided CSS
 * to prevent XSS attacks through CSS injection.
 */

// Allowed CSS variables to prevent XSS attacks
// Only these variables can be set through custom CSS
export const ALLOWED_CSS_VARIABLES = new Set([
  // Primary theme colors (for backward compatibility)
  "--primary",
  "--secondary",
  
  // Core theme colors
  "--ring",
  "--radius",
  "--surface",
  "--background",
  "--foreground",
  
  // UI component colors
  "--card",
  "--card-foreground",
  "--popover",
  "--popover-foreground",
  "--primary-foreground",
  "--secondary-foreground",
  "--muted",
  "--muted-foreground",
  "--accent",
  "--accent-foreground",
  "--destructive",
  "--destructive-foreground",
  "--border",
  "--input",
  
  // Primary theme palette (light mode)
  "--color-theme-50",
  "--color-theme-100",
  "--color-theme-200",
  "--color-theme-300",
  "--color-theme-400",
  "--color-theme-500",
  "--color-theme-600",
  "--color-theme-700",
  "--color-theme-800",
  "--color-theme-900",
  "--color-theme-950",
  
  // Secondary theme palette
  "--color-theme-secondary-50",
  "--color-theme-secondary-100",
  "--color-theme-secondary-200",
  "--color-theme-secondary-300",
  "--color-theme-secondary-400",
  "--color-theme-secondary-500",
  "--color-theme-secondary-600",
  "--color-theme-secondary-700",
  "--color-theme-secondary-800",
  "--color-theme-secondary-900",
  "--color-theme-secondary-950",
  
  // Sequential palette (9 colors)
  "--color-sequential-1",
  "--color-sequential-2",
  "--color-sequential-3",
  "--color-sequential-4",
  "--color-sequential-5",
  "--color-sequential-6",
  "--color-sequential-7",
  "--color-sequential-8",
  "--color-sequential-9",
  
  // Diverging palette (11 colors)
  "--color-diverging-1",
  "--color-diverging-2",
  "--color-diverging-3",
  "--color-diverging-4",
  "--color-diverging-5",
  "--color-diverging-6",
  "--color-diverging-7",
  "--color-diverging-8",
  "--color-diverging-9",
  "--color-diverging-10",
  "--color-diverging-11",
  
  // Qualitative palette (12 colors)
  "--color-qualitative-1",
  "--color-qualitative-2",
  "--color-qualitative-3",
  "--color-qualitative-4",
  "--color-qualitative-5",
  "--color-qualitative-6",
  "--color-qualitative-7",
  "--color-qualitative-8",
  "--color-qualitative-9",
  "--color-qualitative-10",
  "--color-qualitative-11",
  "--color-qualitative-12",
]);

/**
 * Sanitizes CSS input and extracts only known safe CSS variables
 * This prevents XSS attacks through CSS injection
 * 
 * @param css - Raw CSS input from theme
 * @param scopeSelector - Optional CSS selector to scope the variables to (e.g., ".dashboard-theme-boundary")
 * @returns Sanitized CSS containing only safe variable declarations, or null if no safe variables found
 */
export function sanitizeAndExtractSafeVariables(css: string, scopeSelector?: string): string | null {
  const safeVariables = new Map<string, Map<string, string>>();
  
  // Match CSS rules: selector { declarations }
  const rulePattern = /([^{]+)\{([^}]+)\}/g;
  let match: RegExpExecArray | null;
  
  while ((match = rulePattern.exec(css)) !== null) {
    if (!match[1] || !match[2]) continue;
    
    const selector = match[1].trim();
    const declarations = match[2];
    
    // Only allow :root and :root.dark selectors for safety
    if (selector !== ":root" && selector !== ":root.dark" && selector !== ".dark") {
      continue;
    }
    
    // Normalize selector for storage
    const normalizedSelector = selector === ":root" ? ":root" : ":root.dark";
    
    if (!safeVariables.has(normalizedSelector)) {
      safeVariables.set(normalizedSelector, new Map());
    }
    
    // Extract variable declarations: --variable-name: value;
    const varPattern = /--([\w-]+)\s*:\s*([^;]+);?/g;
    let varMatch: RegExpExecArray | null;
    
    while ((varMatch = varPattern.exec(declarations)) !== null) {
      if (!varMatch[1] || !varMatch[2]) continue;
      
      const varName = `--${varMatch[1]}`;
      const varValue = varMatch[2].trim();
      
      // Check if variable is in allowed list
      if (!ALLOWED_CSS_VARIABLES.has(varName)) {
        continue;
      }
      
      // Validate that value is safe (no dangerous patterns)
      if (!isSafeVariableValue(varValue)) {
        console.warn(`Skipping potentially unsafe CSS variable value for ${varName}: ${varValue}`);
        continue;
      }
      
      safeVariables.get(normalizedSelector)!.set(varName, varValue);
    }
  }
  
  // Reconstruct safe CSS
  if (safeVariables.size === 0) {
    return null;
  }
  
  let safeCss = "";
  
  for (const [selector, variables] of safeVariables) {
    if (variables.size === 0) continue;
    
    // Use scoped selector if provided, otherwise use original selector
    const finalSelector = scopeSelector 
      ? (selector === ":root" ? scopeSelector : `${scopeSelector}.dark, .dark ${scopeSelector}`)
      : selector;
    
    safeCss += `${finalSelector} {\n`;
    for (const [name, value] of variables) {
      safeCss += `  ${name}: ${value};\n`;
    }
    safeCss += "}\n\n";
  }
  
  return safeCss.trim() || null;
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

/**
 * Extracts safe CSS variables as a structured object for direct application to elements
 * @param css - Raw CSS input from theme
 * @returns Object with light and dark mode variables, or null if no safe variables found
 */
export function extractSafeVariablesAsObject(css: string): {
  light: Map<string, string>;
  dark: Map<string, string>;
} | null {
  const lightVars = new Map<string, string>();
  const darkVars = new Map<string, string>();
  
  // Match CSS rules: selector { declarations }
  const rulePattern = /([^{]+)\{([^}]+)\}/g;
  let match: RegExpExecArray | null;
  
  while ((match = rulePattern.exec(css)) !== null) {
    if (!match[1] || !match[2]) continue;
    
    const selector = match[1].trim();
    const declarations = match[2];
    
    // Only allow :root and :root.dark selectors for safety
    if (selector !== ":root" && selector !== ":root.dark" && selector !== ".dark") {
      continue;
    }
    
    // Determine if this is dark mode
    const isDark = selector !== ":root";
    const targetMap = isDark ? darkVars : lightVars;
    
    // Extract variable declarations: --variable-name: value;
    const varPattern = /--([\w-]+)\s*:\s*([^;]+);?/g;
    let varMatch: RegExpExecArray | null;
    
    while ((varMatch = varPattern.exec(declarations)) !== null) {
      if (!varMatch[1] || !varMatch[2]) continue;
      
      const varName = `--${varMatch[1]}`;
      const varValue = varMatch[2].trim();
      
      // Check if variable is in allowed list
      if (!ALLOWED_CSS_VARIABLES.has(varName)) {
        continue;
      }
      
      // Validate that value is safe (no dangerous patterns)
      if (!isSafeVariableValue(varValue)) {
        console.warn(`Skipping potentially unsafe CSS variable value for ${varName}: ${varValue}`);
        continue;
      }
      
      targetMap.set(varName, varValue);
    }
  }
  
  // Return null if no variables found
  if (lightVars.size === 0 && darkVars.size === 0) {
    return null;
  }
  
  return { light: lightVars, dark: darkVars };
}

/**
 * Extracts color variables from CSS (--primary and --secondary)
 * Supports both :root.dark and .dark syntax
 */
export function extractColorVariables(css: string): {
  primary: { lightColor: string | null; darkColor: string | null };
  secondary: { lightColor: string | null; darkColor: string | null };
} {
  // Match :root { --primary: ... }
  const lightPrimaryMatch = css.match(/:root\s*\{[^}]*--primary:\s*([^;]+);/s);
  // Match both :root.dark { } and .dark { }
  const darkPrimaryMatch = css.match(/(?::root)?\.dark\s*\{[^}]*--primary:\s*([^;]+);/s);
  const lightSecondaryMatch = css.match(/:root\s*\{[^}]*--secondary:\s*([^;]+);/s);
  const darkSecondaryMatch = css.match(/(?::root)?\.dark\s*\{[^}]*--secondary:\s*([^;]+);/s);
  
  return {
    primary: {
      lightColor: lightPrimaryMatch ? lightPrimaryMatch[1].trim() : null,
      darkColor: darkPrimaryMatch ? darkPrimaryMatch[1].trim() : null,
    },
    secondary: {
      lightColor: lightSecondaryMatch ? lightSecondaryMatch[1].trim() : null,
      darkColor: darkSecondaryMatch ? darkSecondaryMatch[1].trim() : null,
    },
  };
}

