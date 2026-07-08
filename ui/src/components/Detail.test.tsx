import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Detail } from './Detail'
import { WalletDetail, WalletPlanned } from '../utils/model'

const rows: WalletDetail[] = [
  { id: 1, date: 202406, name: 'Coffee', category: 'Daily', currency: 'SGD', amount: -5, done: false, account: 'DBS' },
]
const planned: WalletPlanned = { sgd: 100, idr: 200 }

const handlers = () => ({
  editHandler: jest.fn(),
  deleteHandler: jest.fn(),
  createHandler: jest.fn(),
  updateDashboard: jest.fn(),
})

describe('Detail', () => {
  it('shows the month title and the rows', () => {
    render(<Detail date="202406" rows={rows} planned={planned} {...handlers()} />)
    expect(screen.getByRole('button', { name: 'June 2024' })).toBeInTheDocument()
    expect(screen.getByText('Coffee')).toBeInTheDocument()
  })

  it('navigates to previous / next month', async () => {
    const h = handlers()
    render(<Detail date="202406" rows={rows} planned={planned} {...h} />)

    await userEvent.click(screen.getByTestId('ArrowLeftIcon'))
    expect(h.updateDashboard).toHaveBeenCalledWith('202405')

    await userEvent.click(screen.getByTestId('ArrowRightIcon'))
    expect(h.updateDashboard).toHaveBeenCalledWith('202407')
  })

  it('triggers create / edit / delete', async () => {
    const h = handlers()
    render(<Detail date="202406" rows={rows} planned={planned} {...h} />)

    await userEvent.click(screen.getByTestId('AddIcon'))
    expect(h.createHandler).toHaveBeenCalled()

    await userEvent.click(screen.getByLabelText('edit'))
    expect(h.editHandler).toHaveBeenCalledWith(rows[0])

    await userEvent.click(screen.getByLabelText('delete'))
    expect(h.deleteHandler).toHaveBeenCalledWith(rows[0])
  })

  it('validates the month input in the popover', async () => {
    const h = handlers()
    render(<Detail date="202406" rows={rows} planned={planned} {...h} />)

    await userEvent.click(screen.getByRole('button', { name: 'June 2024' }))
    const popover = await screen.findByRole('presentation')
    const input = within(popover).getByRole('spinbutton') // type=number field

    // Invalid (not 6 digits) -> error, no navigation.
    await userEvent.type(input, '2024')
    await userEvent.click(within(popover).getByRole('button', { name: 'GO!' }))
    expect(within(popover).getByRole('alert')).toHaveTextContent('Salah format goblok!')
    expect(h.updateDashboard).not.toHaveBeenCalled()

    // Valid -> navigates.
    await userEvent.clear(input)
    await userEvent.type(input, '202408')
    await userEvent.click(within(popover).getByRole('button', { name: 'GO!' }))
    expect(h.updateDashboard).toHaveBeenCalledWith('202408')
  })
})


