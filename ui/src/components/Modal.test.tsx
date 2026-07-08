import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { WalletModal } from './Modal'
import { WalletDetail } from '../utils/model'
import { api } from '../utils/api'

jest.mock('../utils/api', () => ({
  api: { post: jest.fn(), delete: jest.fn() },
}))

const mockedApi = api as jest.Mocked<typeof api>

const baseProps = {
  date: '202406',
  onClose: jest.fn(),
  onSuccess: jest.fn(),
}

const existing: WalletDetail = {
  id: 7,
  date: 202406,
  name: 'Old',
  category: 'Daily',
  currency: 'SGD',
  amount: -5,
  done: true,
  account: 'DBS',
}

describe('WalletModal', () => {
  it('creates an entry', async () => {
    mockedApi.post.mockResolvedValue({})
    const onSuccess = jest.fn()
    render(<WalletModal {...baseProps} mode="create" detail={null} onSuccess={onSuccess} />)
    expect(screen.getByRole('heading', { name: 'Create' })).toBeInTheDocument()

    fireEvent.change(document.querySelector('input[name="name"]')!, { target: { value: 'Coffee' } })
    fireEvent.change(document.querySelector('input[name="amount"]')!, { target: { value: '5' } })
    fireEvent.submit(document.querySelector('form')!)

    await waitFor(() =>
      expect(mockedApi.post).toHaveBeenCalledWith(
        '/api/wallet/create',
        expect.objectContaining({ name: 'Coffee', amount: 5, date: 202406 })
      )
    )
    await waitFor(() => expect(onSuccess).toHaveBeenCalled())
  })

  it('updates an entry (includes id, hits /update)', async () => {
    mockedApi.post.mockResolvedValue({})
    render(<WalletModal {...baseProps} mode="edit" detail={existing} />)

    fireEvent.submit(document.querySelector('form')!)

    await waitFor(() =>
      expect(mockedApi.post).toHaveBeenCalledWith(
        '/api/wallet/update',
        expect.objectContaining({ id: 7, name: 'Old' })
      )
    )
  })

  it('deletes an entry', async () => {
    mockedApi.delete.mockResolvedValue({ data: { data: 7 } })
    const onSuccess = jest.fn()
    render(<WalletModal {...baseProps} mode="delete" detail={existing} onSuccess={onSuccess} />)

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))

    await waitFor(() => expect(mockedApi.delete).toHaveBeenCalledWith('/api/wallet/delete/7'))
    await waitFor(() => expect(onSuccess).toHaveBeenCalled())
  })

  it('shows an alert when the request fails', async () => {
    mockedApi.post.mockRejectedValue(new Error('nope'))
    render(<WalletModal {...baseProps} mode="create" detail={null} />)

    fireEvent.submit(document.querySelector('form')!)

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Gagal tot!'))
  })

  it('alerts when delete response id mismatches', async () => {
    mockedApi.delete.mockResolvedValue({ data: { data: 999 } }) // != 7
    const onSuccess = jest.fn()
    render(<WalletModal {...baseProps} mode="delete" detail={existing} onSuccess={onSuccess} />)

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('something is wrong'))
    expect(onSuccess).not.toHaveBeenCalled()
  })

  it('alerts when delete request fails', async () => {
    mockedApi.delete.mockRejectedValue(new Error('boom'))
    render(<WalletModal {...baseProps} mode="delete" detail={existing} />)

    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))

    await waitFor(() => expect(screen.getByRole('alert')).toHaveTextContent('Failed to delete!'))
  })
})


