import { renderHook } from '@testing-library/react'
import { ReactNode } from 'react'
import { useUser } from './useUser'
import { UserContext, UserContextType } from '../UserContext'

describe('useUser', () => {
  it('throws when used outside a provider', () => {
    // Silence the expected React error log for this case.
    const spy = jest.spyOn(console, 'error').mockImplementation(() => {})
    expect(() => renderHook(() => useUser())).toThrow(/UserProvider/)
    spy.mockRestore()
  })

  it('returns the context value inside a provider', () => {
    const value: UserContextType = { userContext: 'tok', saveToken: jest.fn() }
    const wrapper = ({ children }: { children: ReactNode }) => (
      <UserContext.Provider value={value}>{children}</UserContext.Provider>
    )
    const { result } = renderHook(() => useUser(), { wrapper })
    expect(result.current.userContext).toBe('tok')
  })
})

