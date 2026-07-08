import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { StockDashboard } from './StockDashboard'
import { api } from '../utils/api'

jest.mock('../utils/api', () => ({ api: { post: jest.fn() } }))
const mockedApi = api as jest.Mocked<typeof api>

const stocks = [
  { name: 'BBCA', best_price: 100, current_price: 110, fair_price: 200, status: true, buy_price: 100, lot: 2 },
  { name: 'TLKM', best_price: 300, fair_price: 400, status: false },
]

describe('StockDashboard', () => {
  beforeEach(() => {
    mockedApi.post.mockImplementation((url: string) => {
      if (url === '/api/stock/getAll') return Promise.resolve({ data: { data: stocks } })
      if (url === '/api/stock/refresh') return Promise.resolve({ data: { data: stocks } })
      return Promise.resolve({ data: { data: [] } })
    })
  })

  it('loads portfolio and wishlist', async () => {
    render(<StockDashboard />)

    await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/api/stock/getAll', {}))
    expect(screen.getByText('Portfolio')).toBeInTheDocument()
    expect(screen.getByText('Wishlist')).toBeInTheDocument()
    expect(await screen.findByText('BBCA')).toBeInTheDocument()
    expect(screen.getByText('TLKM')).toBeInTheDocument()
  })

  it('refreshes prices', async () => {
    render(<StockDashboard />)
    await screen.findByText('BBCA')

    await userEvent.click(screen.getByRole('button', { name: /Refresh prices/i }))
    await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/api/stock/refresh', {}))
  })

  it('opens the edit modal from a row', async () => {
    render(<StockDashboard />)
    await screen.findByText('BBCA')

    await userEvent.click(screen.getAllByLabelText('edit')[0])
    expect(await screen.findByRole('heading', { name: 'Edit Stock' })).toBeInTheDocument()
  })

  it('shows an alert when getAll fails', async () => {
    mockedApi.post.mockReset()
    mockedApi.post.mockRejectedValue(new Error('down'))
    render(<StockDashboard />)
    expect(await screen.findByRole('alert')).toHaveTextContent('Data failed to fetch/parse!')
  })

  it('shows an alert when refresh fails', async () => {
    mockedApi.post.mockImplementation((url: string) => {
      if (url === '/api/stock/getAll') return Promise.resolve({ data: { data: stocks } })
      return Promise.reject(new Error('refresh down'))
    })
    render(<StockDashboard />)
    await screen.findByText('BBCA')

    await userEvent.click(screen.getByRole('button', { name: /Refresh prices/i }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Failed to refresh prices!')
  })
})

