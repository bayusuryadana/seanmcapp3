import '@testing-library/jest-dom'

// Vite env defaults for tests (import.meta.env is rewritten to process.env by babel).
process.env.MODE ??= 'test'

// jsdom lacks these browser APIs that MUI (useMediaQuery) and recharts rely on.
if (!window.matchMedia) {
  window.matchMedia = (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  }) as unknown as MediaQueryList
}

class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}
const globalWithRO = globalThis as unknown as { ResizeObserver: unknown }
globalWithRO.ResizeObserver = ResizeObserverStub


