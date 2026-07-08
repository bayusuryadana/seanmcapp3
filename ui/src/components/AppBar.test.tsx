import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { WalletAppBar } from './AppBar'

const mockNavigate = jest.fn()
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}))

const renderBar = (logoutHandler = jest.fn()) =>
  render(
    <MemoryRouter>
      <WalletAppBar logoutHandler={logoutHandler} />
    </MemoryRouter>
  )

describe('WalletAppBar', () => {
  it('navigates to the dashboard and stock routes', async () => {
    renderBar()
    await userEvent.click(screen.getByRole('button', { name: 'Dashboard' }))
    expect(mockNavigate).toHaveBeenCalledWith('/wallet')

    await userEvent.click(screen.getByRole('button', { name: 'Stock' }))
    expect(mockNavigate).toHaveBeenCalledWith('/wallet/stock')
  })

  it('calls logoutHandler when the logout icon is clicked', async () => {
    const logout = jest.fn()
    renderBar(logout)
    // The logout button is the icon-only button (no accessible text label).
    await userEvent.click(screen.getByTestId('LogoutIcon'))
    expect(logout).toHaveBeenCalledTimes(1)
  })
})

