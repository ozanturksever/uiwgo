import { describe, it, expect, vi } from 'vitest';

describe('Minimal Test', () => {
  it('should pass a basic test', () => {
    expect(1 + 1).toBe(2);
  });

  it('should have window object in jsdom', () => {
    expect(window).toBeDefined();
    expect(document).toBeDefined();
  });

  it('should have ReactCompatCallbacks mock', () => {
    expect(window.ReactCompatCallbacks).toBeDefined();
    expect(window.ReactCompatCallbacks.invoke).toBeDefined();
    expect(window.ReactCompatCallbacks.register).toBeDefined();
  });
});