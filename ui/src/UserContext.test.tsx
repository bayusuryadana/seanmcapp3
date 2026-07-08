import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { UserProvider } from './UserContext'
import { useUser } from './hooks/useUser'

// A tiny consumer to observe/drive the context.
function Probe() {
  const { userContext, saveToken } = useUser()
  return (
    <div>
      <span data-testid="token">{userContext ?? 'null'}</span>
      <button onClick={() => saveToken('new-token')}>save</button>
      <button onClick={() => saveToken(null)}>logout</button>
    </div>
  )
}

const renderProvider = () =>
  render(
    <UserProvider>
      <Probe />
    </UserProvider>
  )

describe('UserProvider', () => {
  beforeEach(() => localStorage.clear())

  it('loads a valid stored token on mount', () => {
    localStorage.setItem('token', 'stored')
    localStorage.setItem('tokenExpiry', String(Date.now() + 60_000))
    renderProvider()
    expect(screen.getByTestId('token')).toHaveTextContent('stored')
  })

  it('ignores and clears an expired token', () => {
    localStorage.setItem('token', 'stored')
    localStorage.setItem('tokenExpiry', String(Date.now() - 1))
    renderProvider()
    expect(screen.getByTestId('token')).toHaveTextContent('null')
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('saveToken persists token + expiry', async () => {
    renderProvider()
    await userEvent.click(screen.getByText('save'))
    expect(screen.getByTestId('token')).toHaveTextContent('new-token')
    expect(localStorage.getItem('token')).toBe('new-token')
    expect(localStorage.getItem('tokenExpiry')).not.toBeNull()
  })

  it('saveToken(null) clears storage', async () => {
    localStorage.setItem('token', 'stored')
    localStorage.setItem('tokenExpiry', String(Date.now() + 60_000))
    renderProvider()
    await userEvent.click(screen.getByText('logout'))
    expect(screen.getByTestId('token')).toHaveTextContent('null')
    expect(localStorage.getItem('token')).toBeNull()
    expect(localStorage.getItem('tokenExpiry')).toBeNull()
  })
})


