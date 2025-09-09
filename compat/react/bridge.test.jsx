import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import React from 'react';
import { 
  renderComponent, 
  updateComponent, 
  unmountComponent, 
  register, 
  resolve,
  getInstancesMap,
  getComponentRegistry,
  setTheme
} from './bridge.js';

// Test component for bridge testing
function TestComponent({ message = 'Hello Test', onClick }) {
  return (
    <div data-testid="test-message" onClick={onClick}>
      {message}
    </div>
  );
}

// Another test component
function CounterComponent({ count = 0, onIncrement }) {
  return (
    <div>
      <span data-testid="count">{count}</span>
      <button data-testid="increment" onClick={onIncrement}>
        Increment
      </button>
    </div>
  );
}

describe('React Bridge Core', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Clear instances and registry
    const instances = getInstancesMap();
    const registry = getComponentRegistry();
    instances.clear();
    registry.clear();
  });

  afterEach(() => {
    // Clean up any remaining DOM elements
    document.body.innerHTML = '';
  });

  describe('Component Registry', () => {
    it('should register a component', () => {
      register('TestComponent', TestComponent);
      const resolved = resolve('TestComponent');
      expect(resolved).toBe(TestComponent);
    });

    it('should return undefined for unregistered component', () => {
      const resolved = resolve('NonExistentComponent');
      expect(resolved).toBeUndefined();
    });

    it('should overwrite existing component registration', () => {
      register('TestComponent', TestComponent);
      register('TestComponent', CounterComponent);
      const resolved = resolve('TestComponent');
      expect(resolved).toBe(CounterComponent);
    });
  });

  describe('Component Rendering', () => {
    beforeEach(() => {
      register('TestComponent', TestComponent);
      register('CounterComponent', CounterComponent);
    });

    it('should render a component and return instance ID', () => {
      const container = document.createElement('div');
      document.body.appendChild(container);

      const instanceId = renderComponent('TestComponent', { message: 'Hello Bridge' }, container);
      
      expect(instanceId).toBeDefined();
      expect(typeof instanceId).toBe('string');
      expect(container.querySelector('[data-testid="test-message"]')).toHaveTextContent('Hello Bridge');
    });

    it('should throw error for unregistered component', () => {
      const container = document.createElement('div');
      document.body.appendChild(container);

      expect(() => {
        renderComponent('UnknownComponent', {}, container);
      }).toThrow('Component "UnknownComponent" not found in registry');
    });

    it('should store instance in instances map', () => {
      const container = document.createElement('div');
      document.body.appendChild(container);

      const instanceId = renderComponent('TestComponent', { message: 'Test' }, container);
      
      const instances = getInstancesMap();
      expect(instances.has(instanceId)).toBe(true);
      expect(instances.get(instanceId)).toMatchObject({
        componentName: 'TestComponent',
        container: container,
        props: { message: 'Test' }
      });
    });
  });

  describe('Component Updates', () => {
    let container;
    let instanceId;

    beforeEach(() => {
      register('TestComponent', TestComponent);
      container = document.createElement('div');
      document.body.appendChild(container);
      instanceId = renderComponent('TestComponent', { message: 'Initial' }, container);
    });

    it('should update component props', () => {
      updateComponent(instanceId, { message: 'Updated' });
      
      expect(container.querySelector('[data-testid="test-message"]')).toHaveTextContent('Updated');
      
      const instances = getInstancesMap();
      expect(instances.get(instanceId).props).toEqual({ message: 'Updated' });
    });

    it('should throw error for invalid instance ID', () => {
      expect(() => {
        updateComponent('invalid-id', { message: 'Test' });
      }).toThrow('Instance "invalid-id" not found');
    });
  });

  describe('Component Unmounting', () => {
    let container;
    let instanceId;

    beforeEach(() => {
      register('TestComponent', TestComponent);
      container = document.createElement('div');
      document.body.appendChild(container);
      instanceId = renderComponent('TestComponent', { message: 'Test' }, container);
    });

    it('should unmount component and clean up', () => {
      expect(container.children.length).toBeGreaterThan(0);
      
      unmountComponent(instanceId);
      
      expect(container.children.length).toBe(0);
      
      const instances = getInstancesMap();
      expect(instances.has(instanceId)).toBe(false);
    });

    it('should throw error for invalid instance ID', () => {
      expect(() => {
        unmountComponent('invalid-id');
      }).toThrow('Instance "invalid-id" not found');
    });
  });

  describe('ReactCompatCallbacks Integration', () => {
    it('should have ReactCompatCallbacks mock available', () => {
      expect(window.ReactCompatCallbacks).toBeDefined();
      expect(window.ReactCompatCallbacks.invoke).toBeDefined();
      expect(window.ReactCompatCallbacks.register).toBeDefined();
    });
  });

  describe('Theme Management', () => {
    beforeEach(() => {
      // Clear any existing theme classes
      document.documentElement.className = '';
    });

    it('should set dark theme', () => {
      setTheme('dark');
      expect(document.documentElement.classList.contains('dark')).toBe(true);
    });

    it('should set light theme', () => {
      setTheme('light');
      expect(document.documentElement.classList.contains('light')).toBe(true);
    });

    it('should set custom theme with theme- prefix', () => {
      setTheme('custom');
      expect(document.documentElement.classList.contains('theme-custom')).toBe(true);
    });

    it('should replace existing theme classes', () => {
      // Set initial theme
      setTheme('dark');
      expect(document.documentElement.classList.contains('dark')).toBe(true);
      
      // Change to light theme
      setTheme('light');
      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(document.documentElement.classList.contains('light')).toBe(true);
    });

    it('should remove theme classes when theme is empty', () => {
      // Set initial theme
      setTheme('dark');
      expect(document.documentElement.classList.contains('dark')).toBe(true);
      
      // Clear theme
      setTheme('');
      expect(document.documentElement.classList.contains('dark')).toBe(false);
    });

    it('should handle multiple theme class types', () => {
      // Manually add various theme classes
      document.documentElement.classList.add('dark', 'theme-old', 'light');
      
      // Set new theme
      setTheme('theme-new');
      
      // All old theme classes should be removed
      expect(document.documentElement.classList.contains('dark')).toBe(false);
      expect(document.documentElement.classList.contains('light')).toBe(false);
      expect(document.documentElement.classList.contains('theme-old')).toBe(false);
      
      // New theme should be set
      expect(document.documentElement.classList.contains('theme-theme-new')).toBe(true);
    });
  });
});