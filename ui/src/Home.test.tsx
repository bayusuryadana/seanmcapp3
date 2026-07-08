import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { Home } from './Home'

const mockNavigate = jest.fn()
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}))

describe('Home', () => {
  it('navigates to the wallet on button click', async () => {
    render(
      <MemoryRouter>
        <Home />
      </MemoryRouter>
    )
    await userEvent.click(screen.getByRole('button', { name: 'Wallet' }))
    expect(mockNavigate).toHaveBeenCalledWith('/wallet')
  })
})

