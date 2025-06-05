import { browser } from "$app/environment";
import type { V1Conversation } from "../../../runtime-client";

// Storage keys
export const STORAGE_KEYS = {
  CURRENT_CONVERSATION: "rill-current-conversation",
  SIDEBAR_WIDTH: "rill-chat-sidebar-width",
  CHAT_OPEN: "rill-chat-open",
} as const;

// Default values
export const DEFAULTS = {
  SIDEBAR_WIDTH: 500,
  MIN_SIDEBAR_WIDTH: 240,
  MAX_SIDEBAR_WIDTH: 600,
  CHAT_OPEN: false,
} as const;

/**
 * Generic function to get a value from sessionStorage with error handling
 */
function getFromSessionStorage<T>(key: string, defaultValue: T): T {
  if (!browser) return defaultValue;

  try {
    const stored = sessionStorage.getItem(key);
    return stored ? JSON.parse(stored) : defaultValue;
  } catch {
    return defaultValue;
  }
}

/**
 * Generic function to set a value in sessionStorage with error handling
 */
function setInSessionStorage<T>(key: string, value: T): void {
  if (!browser) return;

  try {
    sessionStorage.setItem(key, JSON.stringify(value));
  } catch {
    // Silent fail if sessionStorage is not available
  }
}

/**
 * Generic function to get a value from localStorage with error handling
 */
function getFromLocalStorage<T>(key: string, defaultValue: T): T {
  if (!browser) return defaultValue;

  try {
    const stored = localStorage.getItem(key);
    return stored ? JSON.parse(stored) : defaultValue;
  } catch {
    return defaultValue;
  }
}

/**
 * Generic function to set a value in localStorage with error handling
 */
function setInLocalStorage<T>(key: string, value: T): void {
  if (!browser) return;

  try {
    localStorage.setItem(key, JSON.stringify(value));
  } catch {
    // Silent fail if localStorage is not available
  }
}

/**
 * Remove a value from sessionStorage with error handling
 */
function removeFromSessionStorage(key: string): void {
  if (!browser) return;

  try {
    sessionStorage.removeItem(key);
  } catch {
    // Silent fail
  }
}

// Specific storage functions for chat features

/**
 * Get the current conversation from session storage
 */
export function getCurrentConversation(): V1Conversation | null {
  return getFromSessionStorage(STORAGE_KEYS.CURRENT_CONVERSATION, null);
}

/**
 * Set the current conversation in session storage
 */
export function setCurrentConversation(
  conversation: V1Conversation | null,
): void {
  if (conversation) {
    setInSessionStorage(STORAGE_KEYS.CURRENT_CONVERSATION, conversation);
  } else {
    removeFromSessionStorage(STORAGE_KEYS.CURRENT_CONVERSATION);
  }
}

/**
 * Get the sidebar width from local storage with constraints
 */
export function getSidebarWidth(): number {
  const width = getFromLocalStorage(
    STORAGE_KEYS.SIDEBAR_WIDTH,
    DEFAULTS.SIDEBAR_WIDTH,
  );

  // Ensure width is within bounds
  return Math.max(
    DEFAULTS.MIN_SIDEBAR_WIDTH,
    Math.min(DEFAULTS.MAX_SIDEBAR_WIDTH, width),
  );
}

/**
 * Set the sidebar width in local storage
 */
export function setSidebarWidth(width: number): void {
  setInLocalStorage(STORAGE_KEYS.SIDEBAR_WIDTH, width);
}

/**
 * Get the chat open state from session storage
 */
export function getChatOpenState(): boolean {
  return getFromSessionStorage(STORAGE_KEYS.CHAT_OPEN, DEFAULTS.CHAT_OPEN);
}

/**
 * Set the chat open state in session storage
 */
export function setChatOpenState(isOpen: boolean): void {
  setInSessionStorage(STORAGE_KEYS.CHAT_OPEN, isOpen);
}
