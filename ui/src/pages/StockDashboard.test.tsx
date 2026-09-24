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
      if (url === '/api/stock/summary?period=1d') return Promise.resolve({ data: { data: { jkse: { delta: 100, percentage: 1.43 }, portfolio: { delta: 2000, percentage: 10 }, positions: { BBCA: { delta: 2000, percentage: 10 } } } } })
      return Promise.resolve({ data: { data: [] } })
    })
  })

  it('loads portfolio and wishlist', async () => {
    render(<StockDashboard />)

    await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/api/stock/getAll', {}))
    await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/api/stock/summary?period=1d', {}))
    expect(screen.getByText('Portfolio')).toBeInTheDocument()
    expect(screen.getByText('Wishlist')).toBeInTheDocument()
    expect(await screen.findByText('BBCA')).toBeInTheDocument()
    expect(screen.getByText('TLKM')).toBeInTheDocument()
    expect(screen.getByText('Summary')).toBeInTheDocument()
    expect(screen.getByText('+100 (+1.43%)')).toBeInTheDocument()
    expect(screen.getByText('Rp +2,000 (+10.00%)')).toBeInTheDocument()
  })

  it('collapses and expands the portfolio table', async () => {
    render(<StockDashboard />)
    await screen.findByText('BBCA')

    await userEvent.click(screen.getByRole('button', { name: /Portfolio/i }))
    expect(screen.queryByText('BBCA')).not.toBeVisible()

    await userEvent.click(screen.getByRole('button', { name: /Portfolio/i }))
    expect(screen.getByText('BBCA')).toBeVisible()
  })

  it('refreshes prices', async () => {
    render(<StockDashboard />)
    await screen.findByText('BBCA')

    await userEvent.click(screen.getByRole('button', { name: /Refresh prices/i }))
    await waitFor(() => expect(mockedApi.post).toHaveBeenCalledWith('/api/stock/refresh', {}))
  })

  it('masks and restores the requested money values', async () => {
    render(<StockDashboard />)
    await screen.findByText('BBCA')

    await userEvent.click(screen.getByRole('button', { name: 'Hide money values' }))
    expect(screen.getAllByText('Rp ••••••').length).toBeGreaterThan(0)
    expect(screen.getAllByText('••••••').length).toBeGreaterThan(0)
    expect(screen.queryByText('20,000')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Show money values' }))
    expect(screen.getByText('20,000')).toBeInTheDocument()
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
      if (url === '/api/stock/summary?period=1d') return Promise.resolve({ data: { data: { jkse: { delta: 100, percentage: 1.43 }, portfolio: { delta: 2000, percentage: 10 }, positions: { BBCA: { delta: 2000, percentage: 10 } } } } })
      return Promise.reject(new Error('refresh down'))
    })
    render(<StockDashboard />)
    await screen.findByText('BBCA')

    await userEvent.click(screen.getByRole('button', { name: /Refresh prices/i }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Failed to refresh prices!')
  })
})
