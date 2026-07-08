import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WalletDashboard } from './WalletDashboard'
import { api } from '../utils/api'

jest.mock('../utils/api', () => ({ api: { get: jest.fn(), post: jest.fn() } }))
const mockedApi = api as jest.Mocked<typeof api>

const dashboard = {
  chart: { balance: [{ date: 202406, sum: 5000 }] },
  allocations: [{ name: 'Daily', expense: 10, alloc: 100 }],
  savings: { dbs: 5000, bca: 100000 },
  planned: { sgd: 1, idr: 2 },
  detail: [
    { id: 1, date: 202406, name: 'Coffee', category: 'Daily', currency: 'SGD', amount: -5, done: false, account: 'DBS' },
  ],
}

describe('WalletDashboard', () => {
  it('fetches and renders the dashboard', async () => {
    mockedApi.get.mockResolvedValue({ data: { data: dashboard } })

    render(<WalletDashboard />)

    await waitFor(() =>
      expect(mockedApi.get).toHaveBeenCalledWith(
        '/api/wallet/dashboard',
        expect.objectContaining({ params: expect.objectContaining({ date: expect.any(String) }) })
      )
    )
    expect(await screen.findByText('Coffee')).toBeInTheDocument()
    expect(screen.getByText(/5,000/)).toBeInTheDocument() // DBS savings
  })

  it('shows an alert when the fetch fails', async () => {
    mockedApi.get.mockRejectedValue(new Error('down'))
    render(<WalletDashboard />)
    expect(await screen.findByRole('alert')).toHaveTextContent('Data failed to fetch/parse!')
  })

  it('creates an entry and refetches the dashboard', async () => {
    mockedApi.get.mockResolvedValue({ data: { data: dashboard } })
    mockedApi.post.mockResolvedValue({})
    render(<WalletDashboard />)
    await screen.findByText('Coffee')

    // Open the create modal from the Detail "add" button.
    await userEvent.click(screen.getByTestId('AddIcon'))
    expect(await screen.findByRole('heading', { name: 'Create' })).toBeInTheDocument()

    fireEvent.change(document.querySelector('input[name="name"]')!, { target: { value: 'Tea' } })
    fireEvent.change(document.querySelector('input[name="amount"]')!, { target: { value: '3' } })
    fireEvent.submit(document.querySelector('form')!)

    await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/api/wallet/create', expect.any(Object)))
    // onSuccess triggers a refetch: dashboard loaded once on mount + once after create.
    await waitFor(() => expect(mockedApi.get).toHaveBeenCalledTimes(2))
  })
})

