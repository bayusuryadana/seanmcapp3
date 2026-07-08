import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Stock } from './Stock'
import { WalletStock } from '../utils/model'

const owned: WalletStock = {
  name: 'BBCA',
  best_price: 100,
  current_price: 110,
  fair_price: 200,
  status: true,
  buy_price: 100,
  lot: 2,
}

describe('Stock', () => {
  const handlers = () => ({
    editHandler: jest.fn(),
    deleteHandler: jest.fn(),
    createHandler: jest.fn(),
  })

  it('renders owned columns with computed total bought and P/L', () => {
    render(<Stock title="Portfolio" rows={[owned]} showOwnedColumns {...handlers()} />)

    expect(screen.getByText('Portfolio')).toBeInTheDocument()
    expect(screen.getByText('BBCA')).toBeInTheDocument()
    // total bought = 100 * 2 * 100 = 20,000
    expect(screen.getByText('20,000')).toBeInTheDocument()
    // P/L = (110 - 100) / 100 = +10.00%
    expect(screen.getByText('10.00%')).toBeInTheDocument()
  })

  it('hides owned columns for the wishlist', () => {
    render(
      <Stock
        title="Wishlist"
        rows={[{ name: 'TLKM', best_price: 300, fair_price: 400, status: false }]}
        showOwnedColumns={false}
        {...handlers()}
      />
    )
    expect(screen.queryByText('Total Bought')).not.toBeInTheDocument()
    expect(screen.queryByText('P/L')).not.toBeInTheDocument()
  })

  it('wires up create/edit/delete handlers', async () => {
    const h = handlers()
    render(<Stock title="Portfolio" rows={[owned]} showOwnedColumns {...h} />)

    await userEvent.click(screen.getByTestId('AddIcon'))
    expect(h.createHandler).toHaveBeenCalled()

    await userEvent.click(screen.getByLabelText('edit'))
    expect(h.editHandler).toHaveBeenCalledWith(owned)

    await userEvent.click(screen.getByLabelText('delete'))
    expect(h.deleteHandler).toHaveBeenCalledWith(owned)
  })

  it('renders P/L for loss, break-even and missing price', () => {
    const rows: WalletStock[] = [
      { name: 'LOSS', best_price: 100, current_price: 90, fair_price: 200, status: true, buy_price: 100, lot: 1 }, // -10%
      { name: 'EVEN', best_price: 100, current_price: 100, fair_price: 200, status: true, buy_price: 100, lot: 1 }, // 0%
      { name: 'NONE', best_price: 100, fair_price: 200, status: true }, // no buy/current -> '-'
    ]
    render(<Stock title="Portfolio" rows={rows} showOwnedColumns {...handlers()} />)

    expect(screen.getByText('-10.00%')).toBeInTheDocument()
    expect(screen.getByText('0.00%')).toBeInTheDocument()
    // The row with no prices shows a dash in the P/L cell.
    expect(screen.getAllByText('-').length).toBeGreaterThan(0)
  })
})

