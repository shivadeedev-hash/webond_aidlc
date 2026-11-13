import '@testing-library/jest-dom/vitest';

// jsdom lacks ResizeObserver, which Recharts' ResponsiveContainer relies on.
// Provide a no-op polyfill so chart-bearing pages render under test.
class ResizeObserverStub {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

if (!('ResizeObserver' in globalThis)) {
  (globalThis as unknown as { ResizeObserver: typeof ResizeObserverStub }).ResizeObserver =
    ResizeObserverStub;
}
