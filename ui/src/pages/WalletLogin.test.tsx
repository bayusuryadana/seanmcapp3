import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import axios from 'axios'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { WalletLogin } from './WalletLogin'
import { UserContext, UserContextType } from '../UserContext'
import { api } from '../utils/api'

jest.mock('../utils/api', () => ({ api: { post: jest.fn() } }))
const mockedApi = api as jest.Mocked<typeof api>

function renderLogin(userContext: string | null, saveToken = jest.fn()) {
  const value: UserContextType = { userContext, saveToken }
  render(
    <UserContext.Provider value={value}>
      <MemoryRouter initialEntries={['/wallet/login']}>
        <Routes>
          <Route path="/wallet/login" element={<WalletLogin />} />
          <Route path="/wallet" element={<div>wallet home</div>} />
        </Routes>
      </MemoryRouter>
    </UserContext.Provider>
  )
  return { saveToken }
}

const typePassword = (value: string) =>
  fireEvent.change(document.querySelector('input[name="password"]')!, { target: { value } })

describe('WalletLogin', () => {
  it('logs in and stores the token', async () => {
    mockedApi.post.mockResolvedValue({ data: 'token123' })
    const { saveToken } = renderLogin(null)

    typePassword('secret')
    await userEvent.click(screen.getByRole('button', { name: 'Sign In' }))

    await waitFor(() =>
      expect(mockedApi.post).toHaveBeenCalledWith('/api/wallet/login', { password: 'secret' })
    )
    await waitFor(() => expect(saveToken).toHaveBeenCalledWith('token123'))
  })

  it('shows an error on wrong password (401)', async () => {
    mockedApi.post.mockRejectedValue(
      new axios.AxiosError('unauth', 'ERR', undefined, null, { status: 401 } as never)
    )
    renderLogin(null)

    typePassword('wrong')
    await userEvent.click(screen.getByRole('button', { name: 'Sign In' }))

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Salah password goblok!'))
  })

  it('redirects to /wallet when already authenticated', () => {
    renderLogin('already-a-token')
    expect(screen.getByText('wallet home')).toBeInTheDocument()
  })
})


