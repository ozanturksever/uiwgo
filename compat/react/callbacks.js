/**
 * Callback invocation and event emission for React bridge
 * Integrates with window.ReactCompatCallbacks for Go interop
 */

import { getCallback } from './props.js';
import { handleError, ErrorTypes, createSafeWrapper } from './errors.js';

/**
 * Invoke a callback function by its ID with provided arguments
 * @param {string} callbackId - The callback ID
 * @param {...any} args - Arguments to pass to the callback
 * @returns {any} The callback's return value
 */
export const invoke = createSafeWrapper(function invoke(callbackId, ...args) {
  if (typeof callbackId !== 'string') {
    throw new Error('Callback ID must be a string');
  }

  const callback = getCallback(callbackId);
  if (!callback) {
    console.warn(`Callback with ID ${callbackId} not found`);
    return undefined;
  }

  if (typeof callback !== 'function') {
    console.warn(`Callback with ID ${callbackId} is not a function`);
    return undefined;
  }

  try {
    return callback(...args);
  } catch (error) {
    console.error(`Error invoking callback ${callbackId}:`, error);
    throw error;
  }
}, ErrorTypes.CALLBACK_INVOCATION, { function: 'invoke' });

/**
 * Emit an event to Go via window.ReactCompatCallbacks
 * @param {string} eventType - Type of event to emit
 * @param {any} data - Event data to send
 */
export const emit = createSafeWrapper(function emit(eventType, data = null) {
  if (typeof eventType !== 'string') {
    throw new Error('Event type must be a string');
  }

  // Check if window.ReactCompatCallbacks is available
  if (typeof window === 'undefined' || !window.ReactCompatCallbacks) {
    console.warn('window.ReactCompatCallbacks not available, cannot emit event:', eventType);
    return;
  }

  try {
    // Serialize data for Go consumption
    const serializedData = data !== null ? JSON.stringify(data) : null;
    
    // Call the Go callback
    window.ReactCompatCallbacks.onEvent(eventType, serializedData);
  } catch (error) {
    console.error(`Error emitting event ${eventType}:`, error);
    throw error;
  }
}, ErrorTypes.EVENT_EMISSION, { function: 'emit' });

/**
 * Register error handler for callback invocations
 * @param {Function} handler - Error handler function
 */
let errorHandler = null;

export function onError(handler) {
  if (typeof handler !== 'function') {
    throw new Error('Error handler must be a function');
  }
  errorHandler = handler;
}

/**
 * Get the current error handler
 * @returns {Function|null} Current error handler
 */
export function getErrorHandler() {
  return errorHandler;
}

/**
 * Clear the error handler
 */
export function clearErrorHandler() {
  errorHandler = null;
}

/**
 * Enhanced invoke function with error handling
 * @param {string} callbackId - The callback ID
 * @param {...any} args - Arguments to pass to the callback
 * @returns {any} The callback's return value
 */
export const safeInvoke = createSafeWrapper(function safeInvoke(callbackId, ...args) {
  try {
    return invoke(callbackId, ...args);
  } catch (error) {
    if (errorHandler) {
      try {
        errorHandler(error, callbackId, args);
      } catch (handlerError) {
        console.error('Error in error handler:', handlerError);
      }
    }
    throw error;
  }
}, ErrorTypes.CALLBACK_INVOCATION, { function: 'safeInvoke' });

/**
 * Batch invoke multiple callbacks
 * @param {Array<{id: string, args: Array}>} callbacks - Array of callback invocations
 * @returns {Array} Array of results
 */
export const batchInvoke = createSafeWrapper(function batchInvoke(callbacks) {
  if (!Array.isArray(callbacks)) {
    throw new Error('Callbacks must be an array');
  }

  return callbacks.map(({ id, args = [] }) => {
    try {
      return {
        id,
        success: true,
        result: invoke(id, ...args)
      };
    } catch (error) {
      return {
        id,
        success: false,
        error: error.message
      };
    }
  });
}, ErrorTypes.CALLBACK_INVOCATION, { function: 'batchInvoke' });

/**
 * Initialize the callback system and set up window.ReactCompatCallbacks integration
 */
export function initializeCallbacks() {
  // Ensure window object exists (for browser environment)
  if (typeof window === 'undefined') {
    console.warn('Window object not available, callback system may not work properly');
    return;
  }

  // Initialize ReactCompatCallbacks if not already present
  if (!window.ReactCompatCallbacks) {
    window.ReactCompatCallbacks = {};
  }

  // Expose invoke function to Go
  window.ReactCompatCallbacks.invoke = invoke;
  window.ReactCompatCallbacks.safeInvoke = safeInvoke;
  window.ReactCompatCallbacks.batchInvoke = batchInvoke;
  window.ReactCompatCallbacks.emit = emit;

  // Set up default onEvent handler if not provided
  if (!window.ReactCompatCallbacks.onEvent) {
    window.ReactCompatCallbacks.onEvent = (eventType, data) => {
      console.log(`Event emitted: ${eventType}`, data);
    };
  }

  console.log('ReactCompatCallbacks initialized');
}

/**
 * Check if the callback system is properly initialized
 * @returns {boolean} True if initialized
 */
export function isInitialized() {
  return typeof window !== 'undefined' && 
         !!window.ReactCompatCallbacks && 
         typeof window.ReactCompatCallbacks.invoke === 'function';
}

/**
 * Get callback system status for debugging
 * @returns {Object} Status information
 */
export function getStatus() {
  return {
    windowAvailable: typeof window !== 'undefined',
    reactCompatCallbacksAvailable: typeof window !== 'undefined' && !!window.ReactCompatCallbacks,
    invokeAvailable: typeof window !== 'undefined' && 
                     window.ReactCompatCallbacks && 
                     typeof window.ReactCompatCallbacks.invoke === 'function',
    onEventAvailable: typeof window !== 'undefined' && 
                      window.ReactCompatCallbacks && 
                      typeof window.ReactCompatCallbacks.onEvent === 'function',
    errorHandlerSet: !!errorHandler
  };
}

// Auto-initialize when module is loaded
if (typeof window !== 'undefined') {
  initializeCallbacks();
}