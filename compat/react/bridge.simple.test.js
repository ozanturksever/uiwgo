import { describe, it, expect, beforeEach, vi } from 'vitest';
import React from 'react';
import { 
  register, 
  resolve,
  getComponentRegistry
} from './bridge.js';

// Simple test component
function SimpleComponent({ message = 'Hello' }) {
  return React.createElement('div', { 'data-testid': 'simple-message' }, message);
}

describe('React Bridge - Simple Tests', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    const registry = getComponentRegistry();
    registry.clear();
  });

  describe('Component Registry', () => {
    it('should register a component', () => {
      register('SimpleComponent', SimpleComponent);
      
      const registry = getComponentRegistry();
      expect(registry.has('SimpleComponent')).toBe(true);
      expect(registry.get('SimpleComponent')).toBe(SimpleComponent);
    });

    it('should resolve a registered component', () => {
      register('SimpleComponent', SimpleComponent);
      
      const component = resolve('SimpleComponent');
      expect(component).toBe(SimpleComponent);
    });

    it('should return undefined for unregistered component', () => {
      const component = resolve('UnknownComponent');
      expect(component).toBeUndefined();
    });

    it('should overwrite existing registration', () => {
      const firstComponent = () => React.createElement('div', null, 'First');
      const secondComponent = () => React.createElement('div', null, 'Second');
      
      register('TestComponent', firstComponent);
      register('TestComponent', secondComponent);
      
      expect(resolve('TestComponent')).toBe(secondComponent);
    });
  });

  describe('ReactCompatCallbacks Integration', () => {
    it('should have ReactCompatCallbacks mock available', () => {
      expect(window.ReactCompatCallbacks).toBeDefined();
      expect(window.ReactCompatCallbacks.invoke).toBeDefined();
      expect(window.ReactCompatCallbacks.register).toBeDefined();
    });
  });
});