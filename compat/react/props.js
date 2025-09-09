/**
 * Props serialization and callback handling for React bridge
 * Handles conversion between Go and JS prop formats, including callback placeholders
 */

import { handleError, ErrorTypes, createSafeWrapper } from './errors.js';

// Callback ID counter for generating unique callback identifiers
let callbackIdCounter = 0;

// Map to store callback functions by their IDs
const callbackMap = new Map();

/**
 * Generate a unique callback ID
 * @returns {string} Unique callback identifier
 */
function generateCallbackId() {
  return `cb_${++callbackIdCounter}_${Date.now()}`;
}

/**
 * Deep transform props, converting functions to callback placeholders
 * @param {any} props - Props object to transform
 * @returns {any} Transformed props with callback placeholders
 */
export const serializeProps = createSafeWrapper(function serializeProps(props) {
  if (props === null || props === undefined) {
    return props;
  }

  if (typeof props === 'function') {
    const callbackId = generateCallbackId();
    callbackMap.set(callbackId, props);
    return { __cbid: callbackId };
  }

  if (Array.isArray(props)) {
    return props.map(item => serializeProps(item));
  }

  if (typeof props === 'object') {
    const serialized = {};
    for (const [key, value] of Object.entries(props)) {
      serialized[key] = serializeProps(value);
    }
    return serialized;
  }

  // Primitive values (string, number, boolean) pass through unchanged
  return props;
}, ErrorTypes.PROPS_SERIALIZATION, { function: 'serializeProps' });

/**
 * Deep transform props, converting callback placeholders back to functions
 * @param {any} props - Props object to deserialize
 * @returns {any} Props with callback functions restored
 */
export const deserializeProps = createSafeWrapper(function deserializeProps(props) {
  if (props === null || props === undefined) {
    return props;
  }

  if (Array.isArray(props)) {
    return props.map(item => deserializeProps(item));
  }

  if (typeof props === 'object') {
    // Check if this is a callback placeholder
    if (props.__cbid && typeof props.__cbid === 'string') {
      const callback = callbackMap.get(props.__cbid);
      if (callback) {
        return callback;
      }
      // If callback not found, return a no-op function
      console.warn(`Callback with ID ${props.__cbid} not found`);
      return () => {};
    }

    // Regular object - recursively deserialize
    const deserialized = {};
    for (const [key, value] of Object.entries(props)) {
      deserialized[key] = deserializeProps(value);
    }
    return deserialized;
  }

  // Primitive values pass through unchanged
  return props;
}, ErrorTypes.PROPS_SERIALIZATION, { function: 'deserializeProps' });

/**
 * Store a callback function and return its ID
 * @param {Function} callback - Callback function to store
 * @returns {string} Callback ID
 */
export function storeCallback(callback) {
  if (typeof callback !== 'function') {
    throw new Error('storeCallback expects a function');
  }
  
  const callbackId = generateCallbackId();
  callbackMap.set(callbackId, callback);
  return callbackId;
}

/**
 * Retrieve a callback function by its ID
 * @param {string} callbackId - Callback ID
 * @returns {Function|undefined} The callback function or undefined if not found
 */
export function getCallback(callbackId) {
  return callbackMap.get(callbackId);
}

/**
 * Remove a callback from storage
 * @param {string} callbackId - Callback ID to remove
 * @returns {boolean} True if callback was removed, false if not found
 */
export function removeCallback(callbackId) {
  return callbackMap.delete(callbackId);
}

/**
 * Clear all stored callbacks
 */
export function clearCallbacks() {
  callbackMap.clear();
  callbackIdCounter = 0;
}

/**
 * Get the current callback map (for testing/debugging)
 * @returns {Map} The callback map
 */
export function getCallbackMap() {
  return callbackMap;
}

/**
 * Check if a value is a callback placeholder
 * @param {any} value - Value to check
 * @returns {boolean} True if value is a callback placeholder
 */
export function isCallbackPlaceholder(value) {
  return !!(value && typeof value === 'object' && typeof value.__cbid === 'string');
}

/**
 * Transform props for Go consumption (serialize callbacks)
 * This is the main function Go will call to prepare props
 * @param {any} props - Props to transform
 * @returns {string} JSON string of serialized props
 */
export function transformPropsForGo(props) {
  const serialized = serializeProps(props);
  return JSON.stringify(serialized);
}

/**
 * Transform props from Go (deserialize callbacks)
 * This is the main function to restore props from Go
 * @param {string} propsJson - JSON string of props from Go
 * @returns {any} Deserialized props with callbacks restored
 */
export function transformPropsFromGo(propsJson) {
  try {
    const parsed = JSON.parse(propsJson);
    return deserializeProps(parsed);
  } catch (error) {
    console.error('Failed to parse props from Go:', error);
    return {};
  }
}