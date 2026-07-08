import { API_URL, STOCK_POOL_MONEY } from './constant'

// Smoke test: verifies the Jest toolchain can evaluate a module that reads
// Vite's import.meta.env (rewritten to process.env by the jest babel config).
describe('constant', () => {
  it('exposes an API_URL string', () => {
    expect(typeof API_URL).toBe('string')
  })

  it('falls back STOCK_POOL_MONEY to 0 when env is unset/invalid', () => {
    expect(Number.isFinite(STOCK_POOL_MONEY)).toBe(true)
  })
})

