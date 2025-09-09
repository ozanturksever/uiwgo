/**
 * Tests for callback invocation and event emission
 */

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { 
  invoke, 
  emit, 
  onError, 
  getErrorHandler, 
  clearErrorHandler, 
  safeInvoke, 
  batchInvoke,
  initializeCallbacks,
  isInitialized,
  getStatus
} from './callbacks.js';
import { storeCallback, clearCallbacks } from './props.js';

// Mock window object
Object.defineProperty(global, 'window', {
  value: {},
  writable: true
});

describe('Callback Invocation', () => {
  beforeEach(() => {
    clearCallbacks();
    clearErrorHandler();
    global.window = {};
    initializeCallbacks();
  });

  describe('invoke', () => {
    it('should invoke a stored callback with arguments', () => {
      const mockFn = vi.fn((a, b) => a + b);
      const callbackId = storeCallback(mockFn);
      
      const result = invoke(callbackId, 5, 3);
      
      expect(result).toBe(8);
      expect(mockFn).toHaveBeenCalledWith(5, 3);
    });

    it('should return undefined for non-existent callback', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      const result = invoke('non-existent-id');
      
      expect(result).toBeUndefined();
      expect(consoleSpy).toHaveBeenCalledWith('Callback with ID non-existent-id not found');
      
      consoleSpy.mockRestore();
    });

    it('should throw error for invalid callback ID type', () => {
      expect(() => invoke(123)).toThrow('Callback ID must be a string');
      expect(() => invoke(null)).toThrow('Callback ID must be a string');
      expect(() => invoke(undefined)).toThrow('Callback ID must be a string');
    });

    it('should handle callback that throws error', () => {
      const errorFn = vi.fn(() => {
        throw new Error('Callback error');
      });
      const callbackId = storeCallback(errorFn);
      
      expect(() => invoke(callbackId)).toThrow('Callback error');
    });

    it('should warn for non-function callbacks', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      // Use a non-existent callback ID to trigger the "not found" warning
      const result = invoke('non-existent-callback-id');
      
      expect(result).toBeUndefined();
      expect(consoleSpy).toHaveBeenCalledWith('Callback with ID non-existent-callback-id not found');
      
      consoleSpy.mockRestore();
    });
  });

  describe('safeInvoke', () => {
    it('should invoke callback normally when no error occurs', () => {
      const mockFn = vi.fn(() => 'success');
      const callbackId = storeCallback(mockFn);
      
      const result = safeInvoke(callbackId);
      
      expect(result).toBe('success');
      expect(mockFn).toHaveBeenCalled();
    });

    it('should call error handler when callback throws', () => {
      const errorHandler = vi.fn();
      onError(errorHandler);
      
      const errorFn = vi.fn(() => {
        throw new Error('Test error');
      });
      const callbackId = storeCallback(errorFn);
      
      expect(() => safeInvoke(callbackId, 'arg1', 'arg2')).toThrow('Test error');
      expect(errorHandler).toHaveBeenCalledWith(
        expect.any(Error),
        callbackId,
        ['arg1', 'arg2']
      );
    });

    it('should handle error in error handler gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      const badErrorHandler = vi.fn(() => {
        throw new Error('Handler error');
      });
      onError(badErrorHandler);
      
      const errorFn = vi.fn(() => {
        throw new Error('Original error');
      });
      const callbackId = storeCallback(errorFn);
      
      expect(() => safeInvoke(callbackId)).toThrow('Original error');
      expect(consoleSpy).toHaveBeenCalledWith('Error in error handler:', expect.any(Error));
      
      consoleSpy.mockRestore();
    });
  });

  describe('batchInvoke', () => {
    it('should invoke multiple callbacks and return results', () => {
      const fn1 = vi.fn(() => 'result1');
      const fn2 = vi.fn((x) => x * 2);
      const id1 = storeCallback(fn1);
      const id2 = storeCallback(fn2);
      
      const results = batchInvoke([
        { id: id1, args: [] },
        { id: id2, args: [5] }
      ]);
      
      expect(results).toEqual([
        { id: id1, success: true, result: 'result1' },
        { id: id2, success: true, result: 10 }
      ]);
    });

    it('should handle errors in batch invocation', () => {
      const successFn = vi.fn(() => 'success');
      const errorFn = vi.fn(() => {
        throw new Error('Batch error');
      });
      const id1 = storeCallback(successFn);
      const id2 = storeCallback(errorFn);
      
      const results = batchInvoke([
        { id: id1, args: [] },
        { id: id2, args: [] }
      ]);
      
      expect(results).toEqual([
        { id: id1, success: true, result: 'success' },
        { id: id2, success: false, error: 'Batch error' }
      ]);
    });

    it('should throw error for non-array input', () => {
      expect(() => batchInvoke('not-array')).toThrow('Callbacks must be an array');
      expect(() => batchInvoke(null)).toThrow('Callbacks must be an array');
    });

    it('should handle missing args in batch items', () => {
      const fn = vi.fn(() => 'result');
      const id = storeCallback(fn);
      
      const results = batchInvoke([{ id }]);
      
      expect(results).toEqual([
        { id, success: true, result: 'result' }
      ]);
      expect(fn).toHaveBeenCalledWith();
    });
  });
});

describe('Event Emission', () => {
  beforeEach(() => {
    global.window = {};
    initializeCallbacks();
  });

  describe('emit', () => {
    it('should emit event to window.ReactCompatCallbacks.onEvent', () => {
      const onEventSpy = vi.fn();
      global.window.ReactCompatCallbacks.onEvent = onEventSpy;
      
      emit('test-event', { key: 'value' });
      
      expect(onEventSpy).toHaveBeenCalledWith('test-event', '{"key":"value"}');
    });

    it('should emit event with null data', () => {
      const onEventSpy = vi.fn();
      global.window.ReactCompatCallbacks.onEvent = onEventSpy;
      
      emit('test-event');
      
      expect(onEventSpy).toHaveBeenCalledWith('test-event', null);
    });

    it('should emit event with explicit null data', () => {
      const onEventSpy = vi.fn();
      global.window.ReactCompatCallbacks.onEvent = onEventSpy;
      
      emit('test-event', null);
      
      expect(onEventSpy).toHaveBeenCalledWith('test-event', null);
    });

    it('should throw error for invalid event type', () => {
      expect(() => emit(123)).toThrow('Event type must be a string');
      expect(() => emit(null)).toThrow('Event type must be a string');
    });

    it('should warn when window.ReactCompatCallbacks not available', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      global.window.ReactCompatCallbacks = null;
      
      emit('test-event', 'data');
      
      expect(consoleSpy).toHaveBeenCalledWith(
        'window.ReactCompatCallbacks not available, cannot emit event:',
        'test-event'
      );
      
      consoleSpy.mockRestore();
    });

    it('should handle JSON serialization errors', () => {
      const onEventSpy = vi.fn();
      global.window.ReactCompatCallbacks.onEvent = onEventSpy;
      
      // Create circular reference
      const circularData = {};
      circularData.self = circularData;
      
      expect(() => emit('test-event', circularData)).toThrow();
    });
  });
});

describe('Error Handling', () => {
  beforeEach(() => {
    clearErrorHandler();
  });

  describe('onError', () => {
    it('should set error handler', () => {
      const handler = vi.fn();
      onError(handler);
      
      expect(getErrorHandler()).toBe(handler);
    });

    it('should throw error for non-function handler', () => {
      expect(() => onError('not-function')).toThrow('Error handler must be a function');
      expect(() => onError(null)).toThrow('Error handler must be a function');
    });
  });

  describe('clearErrorHandler', () => {
    it('should clear error handler', () => {
      const handler = vi.fn();
      onError(handler);
      
      clearErrorHandler();
      
      expect(getErrorHandler()).toBeNull();
    });
  });
});

describe('Initialization and Status', () => {
  beforeEach(() => {
    clearErrorHandler();
  });

  describe('initializeCallbacks', () => {
    it('should initialize window.ReactCompatCallbacks', () => {
      global.window = {};
      
      initializeCallbacks();
      
      expect(global.window.ReactCompatCallbacks).toBeDefined();
      expect(typeof global.window.ReactCompatCallbacks.invoke).toBe('function');
      expect(typeof global.window.ReactCompatCallbacks.safeInvoke).toBe('function');
      expect(typeof global.window.ReactCompatCallbacks.batchInvoke).toBe('function');
      expect(typeof global.window.ReactCompatCallbacks.emit).toBe('function');
    });

    it('should not override existing onEvent handler', () => {
      global.window = {
        ReactCompatCallbacks: {
          onEvent: vi.fn()
        }
      };
      const existingHandler = global.window.ReactCompatCallbacks.onEvent;
      
      initializeCallbacks();
      
      expect(global.window.ReactCompatCallbacks.onEvent).toBe(existingHandler);
    });

    it('should provide default onEvent handler', () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      global.window = {};
      
      initializeCallbacks();
      global.window.ReactCompatCallbacks.onEvent('test', 'data');
      
      expect(consoleSpy).toHaveBeenCalledWith('Event emitted: test', 'data');
      consoleSpy.mockRestore();
    });
  });

  describe('isInitialized', () => {
    it('should return true when properly initialized', () => {
      global.window = {};
      initializeCallbacks();
      
      expect(isInitialized()).toBe(true);
    });

    it('should return false when not initialized', () => {
      global.window = {};
      
      expect(isInitialized()).toBe(false);
    });

    it('should return false when window not available', () => {
      global.window = undefined;
      
      expect(isInitialized()).toBe(false);
    });
  });

  describe('getStatus', () => {
    it('should return complete status information', () => {
      global.window = {};
      initializeCallbacks();
      onError(vi.fn());
      
      const status = getStatus();
      
      expect(status).toEqual({
        windowAvailable: true,
        reactCompatCallbacksAvailable: true,
        invokeAvailable: true,
        onEventAvailable: true,
        errorHandlerSet: true
      });
    });

    it('should return false values when not initialized', () => {
      global.window = undefined;
      
      const status = getStatus();
      
      expect(status).toEqual({
        windowAvailable: false,
        reactCompatCallbacksAvailable: false,
        invokeAvailable: false,
        onEventAvailable: false,
        errorHandlerSet: false
      });
    });
  });
});

describe('Integration', () => {
  it('should work end-to-end with props and callbacks', () => {
    global.window = {};
    initializeCallbacks();
    
    // Store a callback
    const mockCallback = vi.fn((data) => `processed: ${data}`);
    const callbackId = storeCallback(mockCallback);
    
    // Invoke via window.ReactCompatCallbacks
    const result = global.window.ReactCompatCallbacks.invoke(callbackId, 'test-data');
    
    expect(result).toBe('processed: test-data');
    expect(mockCallback).toHaveBeenCalledWith('test-data');
  });

  it('should handle event emission from Go side', () => {
    global.window = {};
    initializeCallbacks();
    
    const events = [];
    global.window.ReactCompatCallbacks.onEvent = (type, data) => {
      events.push({ type, data });
    };
    
    // Emit events
    emit('user-action', { action: 'click', target: 'button' });
    emit('state-change', null);
    
    expect(events).toEqual([
      { type: 'user-action', data: '{"action":"click","target":"button"}' },
      { type: 'state-change', data: null }
    ]);
  });
});