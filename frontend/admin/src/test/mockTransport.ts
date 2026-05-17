/**
 * Mock transport for Connect-RPC testing.
 * Use with vi.mock('../lib/transport', () => ({ transport: mockTransport }))
 */

export interface MockCall {
  service: string
  method: string
  request: unknown
}

export function createMockTransport() {
  const calls: MockCall[] = []
  return {
    calls,
    // Call this in tests to set up responses
    _reset() { calls.length = 0 },
    _getCall(idx: number) { return calls[idx] },
    _getCallCount() { return calls.length },
  }
}

export type MockTransport = ReturnType<typeof createMockTransport>
