/**
 * Tests for React bridge error handling and diagnostics
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  ReactBridgeError,
  ErrorTypes,
  ErrorSeverity,
  onError,
  removeErrorHandler,
  clearErrorHandlers,
  getErrorHandlers,
  handleError,
  getErrorHistory,
  clearErrorHistory,
  getErrorStats,
  createSafeWrapper,
  createSafeCatchWrapper,
  createAsyncSafeWrapper,
  createAsyncSafeCatchWrapper,
  getDiagnostics,
  initializeErrorHandling,
  isErrorHandlingInitialized
} from './errors.js';

describe('ReactBridgeError', () => {
  it('should create error with default values', () => {
    const error = new ReactBridgeError('Test error');
    
    expect(error.message).toBe('Test error');
    expect(error.name).toBe('ReactBridgeError');
    expect(error.type).toBe(ErrorTypes.UNKNOWN);
    expect(error.severity).toBe(ErrorSeverity.MEDIUM);
    expect(error.context).toEqual({});
    expect(error.timestamp).toBeDefined();
    expect(error.stack).toBeDefined();
  });

  it('should create error with custom values', () => {
    const context = { componentId: 'test-component' };
    const error = new ReactBridgeError(
      'Component error',
      ErrorTypes.COMPONENT_RENDER,
      ErrorSeverity.HIGH,
      context
    );
    
    expect(error.message).toBe('Component error');
    expect(error.type).toBe(ErrorTypes.COMPONENT_RENDER);
    expect(error.severity).toBe(ErrorSeverity.HIGH);
    expect(error.context).toBe(context);
  });

  it('should serialize to JSON correctly', () => {
    const context = { test: 'value' };
    const error = new ReactBridgeError('Test', ErrorTypes.CALLBACK_INVOCATION, ErrorSeverity.LOW, context);
    const json = error.toJSON();
    
    expect(json).toEqual({
      name: 'ReactBridgeError',
      message: 'Test',
      type: ErrorTypes.CALLBACK_INVOCATION,
      severity: ErrorSeverity.LOW,
      context,
      timestamp: error.timestamp,
      stack: error.stack
    });
  });

  it('should deserialize from JSON correctly', () => {
    const original = new ReactBridgeError('Test', ErrorTypes.PROPS_SERIALIZATION, ErrorSeverity.CRITICAL, { id: 1 });
    const json = original.toJSON();
    const restored = ReactBridgeError.fromJSON(json);
    
    expect(restored.message).toBe(original.message);
    expect(restored.type).toBe(original.type);
    expect(restored.severity).toBe(original.severity);
    expect(restored.context).toEqual(original.context);
    expect(restored.timestamp).toBe(original.timestamp);
  });
});

describe('Error Handler Registration', () => {
  beforeEach(() => {
    clearErrorHandlers();
  });

  it('should register error handler', () => {
    const handler = vi.fn();
    const id = onError(ErrorTypes.COMPONENT_RENDER, handler);
    
    expect(id).toBeDefined();
    expect(getErrorHandlers()).toHaveLength(1);
    expect(getErrorHandlers()[0].types).toEqual([ErrorTypes.COMPONENT_RENDER]);
    expect(getErrorHandlers()[0].handler).toBe(handler);
  });

  it('should register handler with custom ID', () => {
    const handler = vi.fn();
    const customId = 'my-handler';
    const id = onError(ErrorTypes.COMPONENT_RENDER, handler, customId);
    
    expect(id).toBe(customId);
    expect(getErrorHandlers()[0].id).toBe(customId);
  });

  it('should register handler for multiple error types', () => {
    const handler = vi.fn();
    const types = [ErrorTypes.COMPONENT_RENDER, ErrorTypes.COMPONENT_UPDATE];
    onError(types, handler);
    
    expect(getErrorHandlers()[0].types).toEqual(types);
  });

  it('should throw error for non-function handler', () => {
    expect(() => {
      onError(ErrorTypes.COMPONENT_RENDER, 'not a function');
    }).toThrow('Error handler must be a function');
  });

  it('should remove error handler', () => {
    const handler = vi.fn();
    const id = onError(ErrorTypes.COMPONENT_RENDER, handler);
    
    expect(getErrorHandlers()).toHaveLength(1);
    
    const removed = removeErrorHandler(id);
    expect(removed).toBe(true);
    expect(getErrorHandlers()).toHaveLength(0);
  });

  it('should return false when removing non-existent handler', () => {
    const removed = removeErrorHandler('non-existent');
    expect(removed).toBe(false);
  });

  it('should clear all error handlers', () => {
    onError(ErrorTypes.COMPONENT_RENDER, vi.fn());
    onError(ErrorTypes.COMPONENT_UPDATE, vi.fn());
    
    expect(getErrorHandlers()).toHaveLength(2);
    
    clearErrorHandlers();
    expect(getErrorHandlers()).toHaveLength(0);
  });
});

describe('Error Handling', () => {
  beforeEach(() => {
    clearErrorHandlers();
    clearErrorHistory();
    vi.clearAllMocks();
  });

  it('should handle ReactBridgeError', () => {
    const error = new ReactBridgeError('Test error', ErrorTypes.COMPONENT_RENDER);
    const result = handleError(error);
    
    expect(result).toBe(error);
    expect(getErrorHistory()).toHaveLength(1);
  });

  it('should convert regular Error to ReactBridgeError', () => {
    const error = new Error('Regular error');
    const result = handleError(error, ErrorTypes.CALLBACK_INVOCATION);
    
    expect(result).toBeInstanceOf(ReactBridgeError);
    expect(result.message).toBe('Regular error');
    expect(result.type).toBe(ErrorTypes.CALLBACK_INVOCATION);
    expect(result.context.originalError).toBe('Error');
  });

  it('should add additional context to error', () => {
    const error = new Error('Test');
    const additionalContext = { componentId: 'test-123' };
    const result = handleError(error, ErrorTypes.COMPONENT_RENDER, additionalContext);
    
    expect(result.context.componentId).toBe('test-123');
  });

  it('should call registered error handlers', () => {
    const handler1 = vi.fn();
    const handler2 = vi.fn();
    
    onError(ErrorTypes.COMPONENT_RENDER, handler1);
    onError([ErrorTypes.COMPONENT_RENDER, ErrorTypes.COMPONENT_UPDATE], handler2);
    
    const error = new ReactBridgeError('Test', ErrorTypes.COMPONENT_RENDER);
    handleError(error);
    
    expect(handler1).toHaveBeenCalledWith(error);
    expect(handler2).toHaveBeenCalledWith(error);
  });

  it('should call wildcard error handlers', () => {
    const wildcardHandler = vi.fn();
    onError('*', wildcardHandler);
    
    const error = new ReactBridgeError('Test', ErrorTypes.COMPONENT_RENDER);
    handleError(error);
    
    expect(wildcardHandler).toHaveBeenCalledWith(error);
  });

  it('should throw critical errors if not handled', () => {
    const error = new ReactBridgeError('Critical error', ErrorTypes.COMPONENT_RENDER, ErrorSeverity.CRITICAL);
    
    expect(() => {
      handleError(error);
    }).toThrow(error);
  });

  it('should not throw critical errors if handled', () => {
    const handler = vi.fn().mockReturnValue(true); // Mark as handled
    onError(ErrorTypes.COMPONENT_RENDER, handler);
    
    const error = new ReactBridgeError('Critical error', ErrorTypes.COMPONENT_RENDER, ErrorSeverity.CRITICAL);
    
    expect(() => {
      handleError(error);
    }).not.toThrow();
  });

  it('should handle errors in error handlers gracefully', () => {
    const faultyHandler = vi.fn().mockImplementation(() => {
      throw new Error('Handler error');
    });
    onError(ErrorTypes.COMPONENT_RENDER, faultyHandler);
    
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    
    const error = new ReactBridgeError('Test', ErrorTypes.COMPONENT_RENDER);
    
    expect(() => {
      handleError(error);
    }).not.toThrow();
    
    expect(consoleSpy).toHaveBeenCalledWith('Error in error handler:', expect.any(Error));
    
    consoleSpy.mockRestore();
  });


});

describe('Error History', () => {
  beforeEach(() => {
    clearErrorHistory();
  });

  it('should track error history', () => {
    const error1 = new ReactBridgeError('Error 1', ErrorTypes.COMPONENT_RENDER);
    const error2 = new ReactBridgeError('Error 2', ErrorTypes.COMPONENT_UPDATE);
    
    handleError(error1);
    handleError(error2);
    
    const history = getErrorHistory();
    expect(history).toHaveLength(2);
    expect(history[0].message).toBe('Error 1');
    expect(history[1].message).toBe('Error 2');
  });

  it('should limit error history', () => {
    // Create more than 100 errors to test the limit
    for (let i = 0; i < 105; i++) {
      handleError(new Error(`Error ${i}`));
    }
    
    const history = getErrorHistory();
    expect(history).toHaveLength(100); // Should be capped at 100
    expect(history[0].message).toBe('Error 5'); // First 5 should be removed
  });

  it('should return limited history', () => {
    for (let i = 0; i < 10; i++) {
      handleError(new Error(`Error ${i}`));
    }
    
    const history = getErrorHistory(5);
    expect(history).toHaveLength(5);
    expect(history[0].message).toBe('Error 5');
    expect(history[4].message).toBe('Error 9');
  });

  it('should clear error history', () => {
    handleError(new Error('Test'));
    expect(getErrorHistory()).toHaveLength(1);
    
    clearErrorHistory();
    expect(getErrorHistory()).toHaveLength(0);
  });
});

describe('Error Statistics', () => {
  beforeEach(() => {
    clearErrorHistory();
  });

  it('should calculate error statistics', () => {
    handleError(new ReactBridgeError('Error 1', ErrorTypes.COMPONENT_RENDER, ErrorSeverity.HIGH));
    handleError(new ReactBridgeError('Error 2', ErrorTypes.COMPONENT_RENDER, ErrorSeverity.LOW));
    handleError(new ReactBridgeError('Error 3', ErrorTypes.CALLBACK_INVOCATION, ErrorSeverity.HIGH));
    
    const stats = getErrorStats();
    
    expect(stats.total).toBe(3);
    expect(stats.byType[ErrorTypes.COMPONENT_RENDER]).toBe(2);
    expect(stats.byType[ErrorTypes.CALLBACK_INVOCATION]).toBe(1);
    expect(stats.bySeverity[ErrorSeverity.HIGH]).toBe(2);
    expect(stats.bySeverity[ErrorSeverity.LOW]).toBe(1);
    expect(stats.recent).toBe(3); // All recent
  });
});

describe('Safe Wrappers', () => {
    beforeEach(() => {
      clearErrorHistory();
    });

    it('should create safe wrapper that catches errors', () => {
      const errorFunction = () => {
        throw new Error('Function error');
      };
      
      const safeFunction = createSafeCatchWrapper(errorFunction, ErrorTypes.COMPONENT_RENDER);
      
      expect(() => {
        safeFunction();
      }).not.toThrow();
      
      expect(getErrorHistory()).toHaveLength(1);
      expect(getErrorHistory()[0].type).toBe(ErrorTypes.COMPONENT_RENDER);
    });

  it('should create async safe wrapper that catches errors', async () => {
    const faultyAsyncFunction = async () => {
      throw new Error('Async function error');
    };
    
    const safeAsyncFunction = createAsyncSafeCatchWrapper(faultyAsyncFunction, ErrorTypes.CALLBACK_INVOCATION);
    
    await expect(safeAsyncFunction()).resolves.toBeUndefined();
    
    expect(getErrorHistory()).toHaveLength(1);
    expect(getErrorHistory()[0].type).toBe(ErrorTypes.CALLBACK_INVOCATION);
  });

  it('should pass through successful function calls', () => {
    const successfulFunction = (a, b) => a + b;
    const safeFunction = createSafeWrapper(successfulFunction);
    
    const result = safeFunction(2, 3);
    expect(result).toBe(5);
    expect(getErrorHistory()).toHaveLength(0);
  });

  it('should pass through successful async function calls', async () => {
    const successfulAsyncFunction = async (a, b) => a + b;
    const safeAsyncFunction = createAsyncSafeCatchWrapper(successfulAsyncFunction);
    
    const result = await safeAsyncFunction(2, 3);
    expect(result).toBe(5);
    expect(getErrorHistory()).toHaveLength(0);
  });
});

describe('Diagnostics', () => {
  beforeEach(() => {
    clearErrorHistory();
    clearErrorHandlers();
  });

  it('should provide comprehensive diagnostics', () => {
    // Add some errors and handlers
    handleError(new Error('Test error'));
    onError(ErrorTypes.COMPONENT_RENDER, vi.fn());
    
    const diagnostics = getDiagnostics();
    
    expect(diagnostics.timestamp).toBeDefined();
    expect(diagnostics.errorStats.total).toBe(1);
    expect(diagnostics.errorHandlers).toBe(1);
    expect(diagnostics.recentErrors).toHaveLength(1);
    expect(diagnostics.environment).toBeDefined();
    expect(diagnostics.memory.errorHistorySize).toBe(1);
  });
});

describe('Initialization', () => {
  it('should check if error handling is initialized', () => {
    const initialized = isErrorHandlingInitialized();
    expect(initialized).toBe(true); // Should be true in test environment
  });
});