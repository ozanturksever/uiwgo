// Access React and ReactDOM from global scope (UMD)
const ReactGlobal = (typeof globalThis !== 'undefined' ? globalThis.React : undefined) || (typeof window !== 'undefined' ? window.React : undefined);
const ReactDOMGlobal = (typeof globalThis !== 'undefined' ? globalThis.ReactDOM : undefined) || (typeof window !== 'undefined' ? window.ReactDOM : undefined);

console.log('[ReactBridge] ReactGlobal available:', !!ReactGlobal);
console.log('[ReactBridge] ReactDOMGlobal available:', !!ReactDOMGlobal);
console.log('[ReactBridge] window.React available:', !!(typeof window !== 'undefined' && window.React));
console.log('[ReactBridge] window.ReactDOM available:', !!(typeof window !== 'undefined' && window.ReactDOM));

const createRoot = ReactDOMGlobal && ReactDOMGlobal.createRoot ? ReactDOMGlobal.createRoot.bind(ReactDOMGlobal) : null;
const flushSync = ReactDOMGlobal && ReactDOMGlobal.flushSync ? ReactDOMGlobal.flushSync.bind(ReactDOMGlobal) : (fn) => fn();

console.log('[ReactBridge] createRoot available:', !!createRoot);
console.log('[ReactBridge] flushSync available:', !!flushSync);

// Import other modules
import { deserializeProps } from './props.js';
import { initializeCallbacks, invoke, emit } from './callbacks.js';
import { handleError, ErrorTypes, createSafeWrapper } from './errors.js';

// Initialize callback system
initializeCallbacks();

// Component registry
const componentRegistry = new Map();

// Active component instances
const instances = new Map();

// MutationObserver for auto-unmount
let mutationObserver = null;
let observedContainers = new Set();

// Generate unique instance IDs
function generateInstanceId() {
  return `react-instance-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
}

/**
 * Register a React component in the registry
 * @param {string} name - Component name
 * @param {React.Component} component - React component
 */
export const register = createSafeWrapper(function register(name, component) {
  if (typeof component !== 'function') {
    throw new Error(`Component '${name}' must be a function or class`);
  }
  componentRegistry.set(name, component);
  return true;
}, ErrorTypes.BRIDGE_INITIALIZATION, { function: 'register' });

// Bulk registration helper
export const registerMany = createSafeWrapper(function registerMany(components) {
  if (!components || typeof components !== 'object') return;
  for (const [name, component] of Object.entries(components)) {
    if (typeof component === 'function') {
      componentRegistry.set(name, component);
    }
  }
}, ErrorTypes.BRIDGE_INITIALIZATION, { function: 'registerMany' });

/**
 * Resolve a component from the registry
 * @param {string} name - Component name
 * @returns {React.Component|undefined} The component or undefined if not found
 */
export const resolve = createSafeWrapper(function resolve(name) {
  let comp = componentRegistry.get(name);
  if (!comp && typeof window !== 'undefined' && window[name]) {
    comp = window[name];
  }
  return comp;
}, ErrorTypes.BRIDGE_INITIALIZATION, { function: 'resolve' });

/**
 * Render a React component to a DOM container
 * @param {string} componentName - Name of the registered component
 * @param {Object} props - Props to pass to the component
 * @param {HTMLElement|Object} containerOrOptions - DOM container or options object with containerId
 * @returns {string} Instance ID for the rendered component
 */
export const renderComponent = createSafeWrapper(function renderComponent(componentName, props, containerOrOptions) {
  console.log('[ReactBridge] renderComponent called for:', componentName);
  const Component = resolve(componentName);
  
  if (!Component) {
    console.error('[ReactBridge] Component not found:', componentName);
    throw new Error(`Component "${componentName}" not found in registry`);
  }
  
  // Handle both direct container element and options object
  let container;
  if (containerOrOptions && containerOrOptions.containerId) {
    // Options object with containerId
    container = document.getElementById(containerOrOptions.containerId);
    if (!container) {
      throw new Error(`Container with ID "${containerOrOptions.containerId}" not found`);
    }
  } else if (containerOrOptions && containerOrOptions.nodeType) {
    // Direct DOM element
    container = containerOrOptions;
  } else {
    throw new Error('Invalid container: must be a DOM element or options object with containerId');
  }
  
  const instanceId = generateInstanceId();
  if (!createRoot) {
    throw new Error('ReactDOM.createRoot is not available');
  }
  const root = createRoot(container);
  
  // Deserialize props (convert callback placeholders to functions)
  const deserializedProps = deserializeProps(props);

  // Store instance metadata
  instances.set(instanceId, {
    componentName,
    container,
    root,
    props: { ...deserializedProps }
  });

  // Add container to observed containers for auto-unmount
  addObservedContainer(container);

  // Render the component synchronously
  flushSync(() => {
    console.log('[ReactBridge] Rendering with ReactGlobal:', !!ReactGlobal);
    console.log('[ReactBridge] Component function:', Component);
    if (!ReactGlobal) {
      console.error('[ReactBridge] ReactGlobal is not available during render');
      throw new Error('React global not available');
    }
    root.render(ReactGlobal.createElement(Component, deserializedProps));
  });
  
  return instanceId;
}, ErrorTypes.COMPONENT_RENDER, { function: 'renderComponent' });

/**
 * Update props of an existing component instance
 * @param {string} instanceId - Instance ID
 * @param {Object} newProps - New props to apply
 */
export const updateComponent = createSafeWrapper(function updateComponent(instanceId, newProps) {
  const instance = instances.get(instanceId);
  
  if (!instance) {
    throw new Error(`Instance "${instanceId}" not found`);
  }
  
  const Component = resolve(instance.componentName);

  // Deserialize new props
  const deserializedProps = deserializeProps(newProps);

  // Update stored props
  instance.props = { ...deserializedProps };

  // Re-render with new props synchronously
  flushSync(() => {
    if (!ReactGlobal) {
      throw new Error('React global not available');
    }
    instance.root.render(ReactGlobal.createElement(Component, deserializedProps));
  });
}, ErrorTypes.COMPONENT_UPDATE, { function: 'updateComponent' });

/**
 * Unmount a component instance and clean up
 * @param {string} instanceId - Instance ID
 */
export const unmountComponent = createSafeWrapper(function unmountComponent(instanceId) {
  const instance = instances.get(instanceId);
  
  if (!instance) {
    throw new Error(`Instance "${instanceId}" not found`);
  }
  
  // Unmount the React component
  instance.root.unmount();
  
  // Remove from instances map
  instances.delete(instanceId);
  
  // Check if this was the last instance using this container
  const hasOtherInstances = Array.from(instances.values())
    .some(otherInstance => otherInstance.container === instance.container);
  
  // If no other instances use this container, stop observing it
  if (!hasOtherInstances) {
    removeObservedContainer(instance.container);
  }
}, ErrorTypes.COMPONENT_UNMOUNT, { function: 'unmountComponent' });

/**
 * Get the instances map (for testing purposes)
 * @returns {Map} The instances map
 */
export function getInstancesMap() {
  return instances;
}

/**
 * Set the theme by toggling root classes
 * @param {string} theme - Theme name (e.g., 'dark', 'light')
 */
export const setTheme = createSafeWrapper(function setTheme(theme) {
  const documentElement = document.documentElement;
  
  // Remove existing theme classes
  const existingThemeClasses = Array.from(documentElement.classList)
    .filter(className => className.startsWith('theme-') || className === 'dark' || className === 'light');
  
  existingThemeClasses.forEach(className => {
    documentElement.classList.remove(className);
  });
  
  // Add new theme class
  if (theme) {
    // Support both 'dark'/'light' and 'theme-*' patterns
    if (theme === 'dark' || theme === 'light') {
      documentElement.classList.add(theme);
    } else {
      documentElement.classList.add(`theme-${theme}`);
    }
  }
}, ErrorTypes.BRIDGE_INITIALIZATION, { function: 'setTheme' });

/**
 * Get the component registry (for debugging purposes)
 * @returns {Map} The component registry
 */
export function getComponentRegistry() {
  return componentRegistry;
}

/**
 * Initialize MutationObserver for auto-unmount functionality
 */
function initializeMutationObserver() {
  if (mutationObserver || typeof MutationObserver === 'undefined') {
    return;
  }

  mutationObserver = new MutationObserver((mutations) => {
    mutations.forEach((mutation) => {
      if (mutation.type === 'childList') {
        // Check for removed nodes
        mutation.removedNodes.forEach((removedNode) => {
          if (removedNode.nodeType === Node.ELEMENT_NODE) {
            // Check if any of our observed containers were removed
            checkForRemovedContainers(removedNode);
          }
        });
      }
    });
  });

  // Observe the entire document for changes
  mutationObserver.observe(document.body, {
    childList: true,
    subtree: true
  });
}

/**
 * Check if any observed containers were removed and auto-unmount them
 * @param {Element} removedNode - The removed DOM node
 */
function checkForRemovedContainers(removedNode) {
  // Check if the removed node itself is an observed container
  if (observedContainers.has(removedNode)) {
    autoUnmountContainer(removedNode);
  }

  // Check if any observed containers are descendants of the removed node
  observedContainers.forEach((container) => {
    if (removedNode.contains && removedNode.contains(container)) {
      autoUnmountContainer(container);
    }
  });
}

/**
 * Auto-unmount all instances in a removed container
 * @param {Element} container - The removed container
 */
function autoUnmountContainer(container) {
  // Find all instances that use this container
  const instancesToUnmount = [];
  
  instances.forEach((instance, instanceId) => {
    if (instance.container === container) {
      instancesToUnmount.push(instanceId);
    }
  });

  // Unmount all instances in this container
  instancesToUnmount.forEach((instanceId) => {
    try {
      unmountComponent(instanceId);
      console.log(`Auto-unmounted component instance: ${instanceId}`);
    } catch (error) {
      console.warn(`Failed to auto-unmount instance ${instanceId}:`, error);
    }
  });

  // Remove container from observed set
  observedContainers.delete(container);
}

/**
 * Add a container to the observed containers set
 * @param {Element} container - Container to observe
 */
function addObservedContainer(container) {
  observedContainers.add(container);
  
  // Initialize observer if not already done
  if (!mutationObserver) {
    initializeMutationObserver();
  }
}

/**
 * Remove a container from the observed containers set
 * @param {Element} container - Container to stop observing
 */
function removeObservedContainer(container) {
  observedContainers.delete(container);
}

/**
 * Cleanup MutationObserver (for testing purposes)
 */
export function cleanupMutationObserver() {
  if (mutationObserver) {
    mutationObserver.disconnect();
    mutationObserver = null;
  }
  observedContainers.clear();
}

// Expose global ReactCompat object for Go/WASM bridge consumers
const ReactCompatGlobal = {
  renderComponent,
  updateComponent,
  unmountComponent,
  register,
  registerMany,
  resolve,
  setTheme,
  // Diagnostics and utilities
  getDiagnostics: () => ({
    componentsRegistered: getComponentRegistry().size,
    instancesMounted: getInstancesMap().size
  }),
  invoke,
  emit,
};

// Attach to window if available
if (typeof window !== 'undefined') {
  window.ReactCompat = ReactCompatGlobal;
}