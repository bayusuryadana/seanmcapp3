import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WalletStockModal } from './WalletStockModal'
import { WalletStock } from '../utils/model'
import { api } from '../utils/api'

jest.mock('../utils/api', () => ({
  api: { post: jest.fn(), delete: jest.fn() },
}))

const mockedApi = api as jest.Mocked<typeof api>

const baseProps = { onClose: jest.fn(), onSuccess: jest.fn() }

// MUI Modal renders into a portal, so query the whole document.
function fill(name: string, value: string) {
  fireEvent.change(document.querySelector(`input[name="${name}"]`)!, { target: { value } })
}

function submit() {
  fireEvent.submit(document.querySelector('form')!)
}

describe('WalletStockModal', () => {
  it('rejects an invalid stock name', async () => {
    render(
      <WalletStockModal {...baseProps} mode="create" stock={{ name: '', status: false } as WalletStock} />
    )
    fill('name', 'BB') // only 2 letters
    fill('best_price', '100')
    fill('fair_price', '200')
    submit()

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('Name must be exactly 4 capital letters')
    )
    expect(mockedApi.post).not.toHaveBeenCalled()
  })

  it('rejects non-positive prices', async () => {
    render(
      <WalletStockModal {...baseProps} mode="create" stock={{ name: '', status: false } as WalletStock} />
    )
    fill('name', 'BBCA')
    fill('best_price', '0')
    fill('fair_price', '200')
    submit()

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Best Price and Fair Price'))
    expect(mockedApi.post).not.toHaveBeenCalled()
  })

  it('creates a valid stock', async () => {
    mockedApi.post.mockResolvedValue({})
    const onSuccess = jest.fn()
    render(
      <WalletStockModal
        {...baseProps}
        mode="create"
        stock={{ name: '', status: false } as WalletStock}
        onSuccess={onSuccess}
      />
    )
    fill('name', 'BBCA')
    fill('best_price', '100')
    fill('fair_price', '200')
    submit()

    await waitFor(() =>
      expect(mockedApi.post).toHaveBeenCalledWith(
        '/api/stock/create',
        expect.objectContaining({ name: 'BBCA', best_price: 100, fair_price: 200, status: false })
      )
    )
    await waitFor(() => expect(onSuccess).toHaveBeenCalled())
  })

  it('deletes a stock', async () => {
    mockedApi.delete.mockResolvedValue({ data: { data: 'BBCA' } })
    const onSuccess = jest.fn()
    render(
      <WalletStockModal
        {...baseProps}
        mode="delete"
        stock={{ name: 'BBCA', status: true } as WalletStock}
        onSuccess={onSuccess}
      />
    )
    expect(screen.getByRole('alert')).toHaveTextContent('delete')

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))

    await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/api/stock/delete/BBCA'))
    await waitFor(() => expect(onSuccess).toHaveBeenCalled())
  })

  it('creates an owned stock with buy price and lot', async () => {
    mockedApi.post.mockResolvedValue({})
    render(
      <WalletStockModal {...baseProps} mode="create" stock={{ name: '', status: true } as WalletStock} />
    )
    fill('name', 'BBCA')
    fill('best_price', '100')
    fill('fair_price', '200')
    fill('buy_price', '90')
    fill('lot', '3')
    submit()

    await waitFor(() =>
      expect(mockedApi.post).toHaveBeenCalledWith(
        '/api/stock/create',
        expect.objectContaining({ name: 'BBCA', status: true, buy_price: 90, lot: 3 })
      )
    )
  })

  it('rejects owned stock missing buy price / lot', async () => {
    render(
      <WalletStockModal {...baseProps} mode="create" stock={{ name: '', status: true } as WalletStock} />
    )
    fill('name', 'BBCA')
    fill('best_price', '100')
    fill('fair_price', '200')
    submit()

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent('Buy Price and Lot are required')
    )
    expect(mockedApi.post).not.toHaveBeenCalled()
  })

  it('updates an existing stock (hits /update, name disabled)', async () => {
    mockedApi.post.mockResolvedValue({})
    render(
      <WalletStockModal
        {...baseProps}
        mode="edit"
        stock={{ name: 'BBCA', best_price: 100, fair_price: 200, status: false } as WalletStock}
      />
    )
    submit()

    await waitFor(() =>
      expect(mockedApi.post).toHaveBeenCalledWith(
        '/api/stock/update',
        expect.objectContaining({ name: 'BBCA' })
      )
    )
  })
})


