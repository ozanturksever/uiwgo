import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock window.ReactCompatCallbacks for tests
global.window.ReactCompatCallbacks = {
  invoke: vi.fn(),
  register: vi.fn(),
};

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  log: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
};