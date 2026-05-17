// Test setup — runs before each test suite (A13-P2-01)
import '@testing-library/jest-dom'

// Mock ResizeObserver (jsdom doesn't implement it)
;(globalThis as any).ResizeObserver = class {
  observe() {}
  unobserve() {}
  disconnect() {}
}
