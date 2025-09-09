import { describe, it, expect, beforeEach, vi } from 'vitest';
import {
  serializeProps,
  deserializeProps,
  storeCallback,
  getCallback,
  removeCallback,
  clearCallbacks,
  getCallbackMap,
  isCallbackPlaceholder,
  transformPropsForGo,
  transformPropsFromGo
} from './props.js';

describe('Props Serialization and Callback Handling', () => {
  beforeEach(() => {
    clearCallbacks();
  });

  describe('serializeProps', () => {
    it('should handle null and undefined', () => {
      expect(serializeProps(null)).toBe(null);
      expect(serializeProps(undefined)).toBe(undefined);
    });

    it('should handle primitive values', () => {
      expect(serializeProps('hello')).toBe('hello');
      expect(serializeProps(42)).toBe(42);
      expect(serializeProps(true)).toBe(true);
      expect(serializeProps(false)).toBe(false);
    });

    it('should convert functions to callback placeholders', () => {
      const fn = () => console.log('test');
      const result = serializeProps(fn);
      
      expect(result).toHaveProperty('__cbid');
      expect(typeof result.__cbid).toBe('string');
      expect(result.__cbid).toMatch(/^cb_\d+_\d+$/);
    });

    it('should handle arrays with mixed types', () => {
      const fn1 = () => 'fn1';
      const fn2 = () => 'fn2';
      const input = [1, 'hello', fn1, { nested: fn2 }];
      
      const result = serializeProps(input);
      
      expect(result).toHaveLength(4);
      expect(result[0]).toBe(1);
      expect(result[1]).toBe('hello');
      expect(result[2]).toHaveProperty('__cbid');
      expect(result[3].nested).toHaveProperty('__cbid');
    });

    it('should handle nested objects with callbacks', () => {
      const onClick = () => 'clicked';
      const onSubmit = () => 'submitted';
      
      const input = {
        title: 'Test Component',
        count: 42,
        handlers: {
          onClick,
          onSubmit
        },
        config: {
          enabled: true,
          nested: {
            callback: onClick
          }
        }
      };
      
      const result = serializeProps(input);
      
      expect(result.title).toBe('Test Component');
      expect(result.count).toBe(42);
      expect(result.handlers.onClick).toHaveProperty('__cbid');
      expect(result.handlers.onSubmit).toHaveProperty('__cbid');
      expect(result.config.enabled).toBe(true);
      expect(result.config.nested.callback).toHaveProperty('__cbid');
    });
  });

  describe('deserializeProps', () => {
    it('should handle null and undefined', () => {
      expect(deserializeProps(null)).toBe(null);
      expect(deserializeProps(undefined)).toBe(undefined);
    });

    it('should handle primitive values', () => {
      expect(deserializeProps('hello')).toBe('hello');
      expect(deserializeProps(42)).toBe(42);
      expect(deserializeProps(true)).toBe(true);
    });

    it('should convert callback placeholders back to functions', () => {
      const originalFn = vi.fn(() => 'test result');
      const callbackId = storeCallback(originalFn);
      const placeholder = { __cbid: callbackId };
      
      const result = deserializeProps(placeholder);
      
      expect(typeof result).toBe('function');
      expect(result()).toBe('test result');
      expect(originalFn).toHaveBeenCalled();
    });

    it('should handle missing callbacks gracefully', () => {
      const placeholder = { __cbid: 'nonexistent_callback_id' };
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      const result = deserializeProps(placeholder);
      
      expect(typeof result).toBe('function');
      expect(result()).toBe(undefined); // no-op function
      expect(consoleSpy).toHaveBeenCalledWith('Callback with ID nonexistent_callback_id not found');
      
      consoleSpy.mockRestore();
    });

    it('should handle nested objects with callback placeholders', () => {
      const onClick = vi.fn(() => 'clicked');
      const onSubmit = vi.fn(() => 'submitted');
      
      const clickId = storeCallback(onClick);
      const submitId = storeCallback(onSubmit);
      
      const input = {
        title: 'Test',
        handlers: {
          onClick: { __cbid: clickId },
          onSubmit: { __cbid: submitId }
        }
      };
      
      const result = deserializeProps(input);
      
      expect(result.title).toBe('Test');
      expect(typeof result.handlers.onClick).toBe('function');
      expect(typeof result.handlers.onSubmit).toBe('function');
      expect(result.handlers.onClick()).toBe('clicked');
      expect(result.handlers.onSubmit()).toBe('submitted');
    });
  });

  describe('callback storage and retrieval', () => {
    it('should store and retrieve callbacks', () => {
      const fn = vi.fn(() => 'test');
      const callbackId = storeCallback(fn);
      
      expect(typeof callbackId).toBe('string');
      expect(callbackId).toMatch(/^cb_\d+_\d+$/);
      
      const retrieved = getCallback(callbackId);
      expect(retrieved).toBe(fn);
      expect(retrieved()).toBe('test');
    });

    it('should throw error when storing non-function', () => {
      expect(() => storeCallback('not a function')).toThrow('storeCallback expects a function');
      expect(() => storeCallback(42)).toThrow('storeCallback expects a function');
      expect(() => storeCallback({})).toThrow('storeCallback expects a function');
    });

    it('should remove callbacks', () => {
      const fn = () => 'test';
      const callbackId = storeCallback(fn);
      
      expect(getCallback(callbackId)).toBe(fn);
      expect(removeCallback(callbackId)).toBe(true);
      expect(getCallback(callbackId)).toBe(undefined);
      expect(removeCallback(callbackId)).toBe(false); // already removed
    });

    it('should clear all callbacks', () => {
      const fn1 = () => 'fn1';
      const fn2 = () => 'fn2';
      
      const id1 = storeCallback(fn1);
      const id2 = storeCallback(fn2);
      
      expect(getCallbackMap().size).toBe(2);
      
      clearCallbacks();
      
      expect(getCallbackMap().size).toBe(0);
      expect(getCallback(id1)).toBe(undefined);
      expect(getCallback(id2)).toBe(undefined);
    });
  });

  describe('utility functions', () => {
    it('should identify callback placeholders', () => {
      expect(isCallbackPlaceholder({ __cbid: 'test_id' })).toBe(true);
      expect(isCallbackPlaceholder({ __cbid: 123 })).toBe(false); // number, not string
      expect(isCallbackPlaceholder({ other: 'value' })).toBe(false);
      expect(isCallbackPlaceholder('string')).toBe(false);
      expect(isCallbackPlaceholder(null)).toBe(false);
      expect(isCallbackPlaceholder(undefined)).toBe(false);
    });
  });

  describe('Go integration functions', () => {
    it('should transform props for Go consumption', () => {
      const onClick = () => 'clicked';
      const props = {
        title: 'Test',
        count: 42,
        onClick
      };
      
      const result = transformPropsForGo(props);
      
      expect(typeof result).toBe('string');
      const parsed = JSON.parse(result);
      expect(parsed.title).toBe('Test');
      expect(parsed.count).toBe(42);
      expect(parsed.onClick).toHaveProperty('__cbid');
    });

    it('should transform props from Go', () => {
      const onClick = vi.fn(() => 'clicked');
      const callbackId = storeCallback(onClick);
      
      const propsJson = JSON.stringify({
        title: 'Test',
        count: 42,
        onClick: { __cbid: callbackId }
      });
      
      const result = transformPropsFromGo(propsJson);
      
      expect(result.title).toBe('Test');
      expect(result.count).toBe(42);
      expect(typeof result.onClick).toBe('function');
      expect(result.onClick()).toBe('clicked');
    });

    it('should handle invalid JSON gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      
      const result = transformPropsFromGo('invalid json {');
      
      expect(result).toEqual({});
      expect(consoleSpy).toHaveBeenCalledWith('Failed to parse props from Go:', expect.any(Error));
      
      consoleSpy.mockRestore();
    });
  });

  describe('round-trip serialization', () => {
    it('should maintain data integrity through serialize/deserialize cycle', () => {
      const onClick = vi.fn(() => 'clicked');
      const onSubmit = vi.fn(() => 'submitted');
      
      const originalProps = {
        title: 'Test Component',
        count: 42,
        enabled: true,
        items: ['a', 'b', 'c'],
        handlers: {
          onClick,
          onSubmit
        },
        nested: {
          deep: {
            callback: onClick
          }
        }
      };
      
      // Serialize then deserialize
      const serialized = serializeProps(originalProps);
      const deserialized = deserializeProps(serialized);
      
      // Check primitive values
      expect(deserialized.title).toBe(originalProps.title);
      expect(deserialized.count).toBe(originalProps.count);
      expect(deserialized.enabled).toBe(originalProps.enabled);
      expect(deserialized.items).toEqual(originalProps.items);
      
      // Check callbacks work
      expect(typeof deserialized.handlers.onClick).toBe('function');
      expect(typeof deserialized.handlers.onSubmit).toBe('function');
      expect(typeof deserialized.nested.deep.callback).toBe('function');
      
      expect(deserialized.handlers.onClick()).toBe('clicked');
      expect(deserialized.handlers.onSubmit()).toBe('submitted');
      expect(deserialized.nested.deep.callback()).toBe('clicked');
    });
  });
});