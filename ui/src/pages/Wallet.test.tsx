import { render, screen } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { Wallet } from './Wallet'
import { UserContext, UserContextType } from '../UserContext'

function renderWallet(userContext: string | null) {
  const value: UserContextType = { userContext, saveToken: jest.fn() }
  return render(
    <UserContext.Provider value={value}>
      <MemoryRouter initialEntries={['/wallet']}>
        <Routes>
          <Route path="/wallet" element={<Wallet />}>
            <Route index element={<div>dashboard content</div>} />
          </Route>
          <Route path="/wallet/login" element={<div>login page</div>} />
        </Routes>
      </MemoryRouter>
    </UserContext.Provider>
  )
}

describe('Wallet layout', () => {
  it('renders the app bar and outlet when authenticated', () => {
    renderWallet('a-token')
    expect(screen.getByText('Seanmcwallet')).toBeInTheDocument()
    expect(screen.getByText('dashboard content')).toBeInTheDocument()
  })

  it('redirects to login when not authenticated', () => {
    renderWallet(null)
    expect(screen.getByText('login page')).toBeInTheDocument()
  })
})

