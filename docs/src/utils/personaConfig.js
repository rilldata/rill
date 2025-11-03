/**
 * Persona Switcher Configuration
 * 
 * This file defines which documentation sections are visible for different user personas.
 * The PersonaSwitcher component in the navbar allows users to toggle between:
 * - Developer: Technical users who need build, deploy, and integration docs
 * - Business User: Non-technical users focused on exploring and managing dashboards
 * 
 * The selection is stored in localStorage and persists across sessions.
 * When changed, the sidebar automatically filters to show only relevant sections.
 * 
 * To modify which sections appear for each persona, update the 'visible' arrays below.
 */

// Configuration for which pages/sections are visible for each persona
// Supports:
// - Wildcards: 'get-started/*' (all pages under get-started)
// - Specific pages: 'build/connectors/credentials' (exact page only)
// - Section shorthand: 'reference' (automatically becomes 'reference/*')
//
// Mix and match as needed! Examples:
// - Show all build pages: 'build/*'
// - Show only one build page: 'build/connectors/credentials'
// - Show a subsection: 'guides/rill-basics/*'
// - Show specific page in subsection: 'explore/dashboard-101/pivot'
export const PERSONA_CONFIG = {
    developer: {
        visible: [
            'get-started/*',           // All get-started pages
            'build/*',                 // All build pages
            'deploy/*',                // All deploy pages
            'integrate/*',             // All integrate pages
            'guides/*',                // All guides
            'other/*',                 // All other pages
            'reference/*',             // All reference pages
            'contact/*'                // Contact pages
        ],
        label: 'Developer'
    },
    business: {
        visible: [
            'get-started/*',           // All get-started pages
            'explore/*',               // All explore pages
            'manage/*',                // All manage pages
            'guides/*',                // All guides
            'contact/*'                // Contact pages
        ],
        label: 'Business User'
    }
};

/**
 * Extract the full path from various item properties
 */
function extractItemPath(item) {
    // Try different properties to get the full path
    if (item.docId) {
        return item.docId;
    }

    if (item.id) {
        return item.id;
    }

    if (item.href) {
        // Remove leading slash and extract path
        const match = item.href.match(/^\/(.+)$/);
        if (match) return match[1];
    }

    if (item.link?.id) {
        return item.link.id;
    }

    if (item.link?.href) {
        const match = item.link.href.match(/^\/(.+)$/);
        if (match) return match[1];
    }

    // For category items, check the first child item
    if (item.items && item.items.length > 0) {
        return extractItemPath(item.items[0]);
    }

    // Last resort: try to match by label
    if (item.label) {
        const normalized = item.label.toLowerCase().replace(/\s+/g, '-');
        return normalized;
    }

    return null;
}

/**
 * Check if a path matches a pattern (supports wildcards)
 * Examples:
 * - matchesPattern('build/connectors/s3', 'build/*') => true
 * - matchesPattern('build/connectors/s3', 'build/connectors/s3') => true
 * - matchesPattern('explore/filters', 'build/*') => false
 */
function matchesPattern(path, pattern) {
    if (!path || !pattern) return false;

    // Normalize: if pattern doesn't have wildcard or slash, treat as section wildcard
    // e.g., 'reference' becomes 'reference/*'
    let normalizedPattern = pattern;
    if (!pattern.includes('*') && !pattern.includes('/')) {
        normalizedPattern = `${pattern}/*`;
    }

    // Exact match
    if (path === normalizedPattern) return true;

    // Wildcard match
    if (normalizedPattern.endsWith('/*')) {
        const prefix = normalizedPattern.slice(0, -2); // Remove '/*'
        return path === prefix || path.startsWith(prefix + '/');
    }

    // Check if path starts with the pattern
    if (normalizedPattern.endsWith('*')) {
        const prefix = normalizedPattern.slice(0, -1);
        return path.startsWith(prefix);
    }

    return false;
}

/**
 * Check if a sidebar item should be visible based on current persona
 */
export function isItemVisibleForPersona(item, persona) {
    if (!item || !persona) return true;

    const config = PERSONA_CONFIG[persona];
    if (!config) return true;

    const itemPath = extractItemPath(item);

    // If we have a path, check if it matches any visible patterns
    if (itemPath) {
        return config.visible.some(pattern => matchesPattern(itemPath, pattern));
    }

    // If we can't determine the path, show it by default
    return true;
}

/**
 * Filter sidebar items based on persona
 */
export function filterSidebarItems(items, persona) {
    if (!items || !Array.isArray(items)) return items;

    return items
        .filter(item => isItemVisibleForPersona(item, persona))
        .map(item => {
            // If item has children, filter them recursively
            if (item.items && Array.isArray(item.items)) {
                return {
                    ...item,
                    items: filterSidebarItems(item.items, persona)
                };
            }
            return item;
        });
}

