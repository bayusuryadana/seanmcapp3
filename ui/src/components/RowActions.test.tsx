import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { RowActions } from './RowActions'

function renderInTable(ui: React.ReactNode) {
  return render(
    <table>
      <tbody>
        <tr>{ui}</tr>
      </tbody>
    </table>
  )
}

describe('RowActions', () => {
  it('calls onEdit and onDelete', async () => {
    const onEdit = jest.fn()
    const onDelete = jest.fn()
    renderInTable(<RowActions onEdit={onEdit} onDelete={onDelete} />)

    await userEvent.click(screen.getByLabelText('edit'))
    expect(onEdit).toHaveBeenCalledTimes(1)

    await userEvent.click(screen.getByLabelText('delete'))
    expect(onDelete).toHaveBeenCalledTimes(1)
  })
})

