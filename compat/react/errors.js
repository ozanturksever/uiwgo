/**
 * Error handling and diagnostics for React bridge
 * Provides comprehensive error tracking, logging, and recovery
 */

/**
 * Error types for categorization
 */
export const ErrorTypes = {
  COMPONENT_RENDER: 'component_render',
  COMPONENT_UPDATE: 'component_update', 
  COMPONENT_UNMOUNT: 'component_unmount',
  CALLBACK_INVOCATION: 'callback_invocation',
  EVENT_EMISSION: 'event_emission',
  PROPS_SERIALIZATION: 'props_serialization',
  BRIDGE_INITIALIZATION: 'bridge_initialization',
  DOM_MANIPULATION: 'dom_manipulation',
  UNKNOWN: 'unknown'
};

/**
 * Error severity levels
 */
export const ErrorSeverity = {
  LOW: 'low',
  MEDIUM: 'medium',
  HIGH: 'high',
  CRITICAL: 'critical'
};

/**
 * Global error handlers registry
 */
const errorHandlers = new Map();
const errorHistory = [];
const maxErrorHistory = 100;

/**
 * Enhanced error class with additional context
 */
export class ReactBridgeError extends Error {
  constructor(message, type = ErrorTypes.UNKNOWN, severity = ErrorSeverity.MEDIUM, context = {}) {
    super(message);
    this.name = 'ReactBridgeError';
    this.type = type;
    this.severity = severity;
    this.context = context;
    this.timestamp = new Date().toISOString();
    this.stack = this.stack || new Error().stack;
  }

  /**
   * Convert error to serializable object
   * @returns {Object} Serializable error representation
   */
  toJSON() {
    return {
      name: this.name,
      message: this.message,
      type: this.type,
      severity: this.severity,
      context: this.context,
      timestamp: this.timestamp,
      stack: this.stack
    };
  }

  /**
   * Create error from serialized object
   * @param {Object} obj - Serialized error object
   * @returns {ReactBridgeError} Reconstructed error
   */
  static fromJSON(obj) {
    const error = new ReactBridgeError(obj.message, obj.type, obj.severity, obj.context);
    error.timestamp = obj.timestamp;
    error.stack = obj.stack;
    return error;
  }
}

/**
 * Register an error handler for specific error types
 * @param {string|Array<string>} types - Error type(s) to handle
 * @param {Function} handler - Error handler function
 * @param {string} id - Unique identifier for the handler
 */
export function onError(types, handler, id = null) {
  if (typeof handler !== 'function') {
    throw new Error('Error handler must be a function');
  }

  const typeArray = Array.isArray(types) ? types : [types];
  const handlerId = id || `handler_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

  errorHandlers.set(handlerId, {
    types: typeArray,
    handler,
    id: handlerId
  });

  return handlerId;
}

/**
 * Remove an error handler
 * @param {string} handlerId - Handler ID to remove
 * @returns {boolean} True if handler was removed
 */
export function removeErrorHandler(handlerId) {
  return errorHandlers.delete(handlerId);
}

/**
 * Clear all error handlers
 */
export function clearErrorHandlers() {
  errorHandlers.clear();
}

/**
 * Get all registered error handlers
 * @returns {Array} Array of handler information
 */
export function getErrorHandlers() {
  return Array.from(errorHandlers.values());
}

/**
 * Handle an error through the error handling system
 * @param {Error|ReactBridgeError} error - Error to handle
 * @param {string} type - Error type override
 * @param {Object} additionalContext - Additional context to merge
 */
export function handleError(error, type = null, additionalContext = {}) {
  // Convert to ReactBridgeError if needed
  let bridgeError;
  if (error instanceof ReactBridgeError) {
    bridgeError = error;
    if (type) bridgeError.type = type;
    bridgeError.context = { ...bridgeError.context, ...additionalContext };
  } else {
    bridgeError = new ReactBridgeError(
      error.message || 'Unknown error',
      type || ErrorTypes.UNKNOWN,
      ErrorSeverity.MEDIUM,
      {
        originalError: error.name || 'Error',
        stack: error.stack,
        ...additionalContext
      }
    );
  }

  // Add to error history
  errorHistory.push(bridgeError.toJSON());
  if (errorHistory.length > maxErrorHistory) {
    errorHistory.shift();
  }

  // Log error
  logError(bridgeError);

  // Note: Error event emission would be handled by external systems
  // to avoid circular dependencies

  // Call registered handlers
  let handled = false;
  for (const { types, handler } of errorHandlers.values()) {
    if (types.includes(bridgeError.type) || types.includes('*')) {
      try {
        const result = handler(bridgeError);
        if (result === true) {
          handled = true;
        }
      } catch (handlerError) {
        console.error('Error in error handler:', handlerError);
      }
    }
  }

  // If no handler marked the error as handled, and it's critical, throw it
  if (!handled && bridgeError.severity === ErrorSeverity.CRITICAL) {
    throw bridgeError;
  }

  return bridgeError;
}

/**
 * Log error with appropriate formatting
 * @param {ReactBridgeError} error - Error to log
 */
function logError(error) {
  const prefix = `[ReactBridge:${error.type}:${error.severity}]`;
  const message = `${prefix} ${error.message}`;
  
  switch (error.severity) {
    case ErrorSeverity.CRITICAL:
      console.error(message, error.context, error.stack);
      break;
    case ErrorSeverity.HIGH:
      console.error(message, error.context);
      break;
    case ErrorSeverity.MEDIUM:
      console.warn(message, error.context);
      break;
    case ErrorSeverity.LOW:
      console.log(message, error.context);
      break;
    default:
      console.log(message, error.context);
  }
}

/**
 * Get error history
 * @param {number} limit - Maximum number of errors to return (default: all)
 * @returns {Array} Array of serialized errors
 */
export function getErrorHistory(limit = null) {
  if (limit === null) {
    return [...errorHistory];
  }
  const startIndex = Math.max(0, errorHistory.length - limit);
  return errorHistory.slice(startIndex);
}

/**
 * Clear error history
 */
export function clearErrorHistory() {
  errorHistory.length = 0;
}

/**
 * Get error statistics
 * @returns {Object} Error statistics
 */
export function getErrorStats() {
  const stats = {
    total: errorHistory.length,
    byType: {},
    bySeverity: {},
    recent: 0 // Last 5 minutes
  };

  const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString();

  for (const error of errorHistory) {
    // Count by type
    stats.byType[error.type] = (stats.byType[error.type] || 0) + 1;
    
    // Count by severity
    stats.bySeverity[error.severity] = (stats.bySeverity[error.severity] || 0) + 1;
    
    // Count recent errors
    if (error.timestamp > fiveMinutesAgo) {
      stats.recent++;
    }
  }

  return stats;
}

/**
 * Create a safe wrapper function that handles errors
 * @param {Function} fn - Function to wrap
 * @param {string} errorType - Error type for any thrown errors
 * @param {Object} context - Additional context for errors
 * @returns {Function} Wrapped function
 */
export function createSafeWrapper(fn, errorType = ErrorTypes.UNKNOWN, context = {}) {
  return function(...args) {
    try {
      return fn.apply(this, args);
    } catch (error) {
      handleError(error, errorType, {
        functionName: fn.name || 'anonymous',
        arguments: args.length,
        ...context
      });
      // Re-throw the error to maintain expected behavior
      throw error;
    }
  };
}

/**
 * Create an async safe wrapper function that handles errors
 * @param {Function} fn - Async function to wrap
 * @param {string} errorType - Error type for any thrown errors
 * @param {Object} context - Additional context for errors
 * @returns {Function} Wrapped async function
 */
export function createAsyncSafeWrapper(fn, errorType = ErrorTypes.UNKNOWN, context = {}) {
  return async function(...args) {
    try {
      return await fn.apply(this, args);
    } catch (error) {
      handleError(error, errorType, {
        functionName: fn.name || 'anonymous',
        arguments: args.length,
        async: true,
        ...context
      });
      // Re-throw the error to maintain expected behavior
      throw error;
    }
  };
}

/**
 * Create a safe wrapper function that catches and handles errors without re-throwing
 * @param {Function} fn - Function to wrap
 * @param {string} errorType - Error type for any thrown errors
 * @param {Object} context - Additional context for errors
 * @returns {Function} Wrapped function
 */
export function createSafeCatchWrapper(fn, errorType = ErrorTypes.UNKNOWN, context = {}) {
  return function(...args) {
    try {
      return fn.apply(this, args);
    } catch (error) {
      handleError(error, errorType, {
        functionName: fn.name || 'anonymous',
        arguments: args.length,
        async: true,
        ...context
      });
      return undefined; // Return undefined instead of re-throwing
    }
  };
}

/**
 * Create an async safe wrapper function that catches and handles errors without re-throwing
 * @param {Function} fn - Async function to wrap
 * @param {string} errorType - Error type for any thrown errors
 * @param {Object} context - Additional context for errors
 * @returns {Function} Wrapped async function
 */
export function createAsyncSafeCatchWrapper(fn, errorType = ErrorTypes.UNKNOWN, context = {}) {
  return async function(...args) {
    try {
      return await fn.apply(this, args);
    } catch (error) {
      handleError(error, errorType, {
        functionName: fn.name || 'anonymous',
        arguments: args.length,
        async: true,
        ...context
      });
    }
  };
}

/**
 * Diagnostic information collector
 * @returns {Object} Comprehensive diagnostic information
 */
export function getDiagnostics() {
  return {
    timestamp: new Date().toISOString(),
    errorStats: getErrorStats(),
    errorHandlers: getErrorHandlers().length,
    recentErrors: getErrorHistory(10),
    environment: {
      userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : 'unknown',
      windowAvailable: typeof window !== 'undefined',
      reactCompatCallbacks: typeof window !== 'undefined' && !!window.ReactCompatCallbacks
    },
    memory: {
      errorHistorySize: errorHistory.length,
      maxErrorHistory
    }
  };
}

/**
 * Initialize error handling system
 */
export function initializeErrorHandling() {
  // Set up global error handler for unhandled errors
  if (typeof window !== 'undefined') {
    window.addEventListener('error', (event) => {
      handleError(event.error || new Error(event.message), ErrorTypes.UNKNOWN, {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        global: true
      });
    });

    window.addEventListener('unhandledrejection', (event) => {
      handleError(event.reason || new Error('Unhandled promise rejection'), ErrorTypes.UNKNOWN, {
        promise: true,
        global: true
      });
    });
  }

  console.log('ReactBridge error handling initialized');
}

/**
 * Check if error handling is properly initialized
 * @returns {boolean} True if initialized
 */
export function isErrorHandlingInitialized() {
  return typeof window !== 'undefined' && 
         window.addEventListener !== undefined;
}

// Auto-initialize when module is loaded
if (typeof window !== 'undefined') {
  initializeErrorHandling();
}