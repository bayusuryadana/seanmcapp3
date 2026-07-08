import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { FormModal } from './FormModal'

describe('FormModal', () => {
  const baseProps = {
    open: true,
    title: 'Create',
    submitLabel: 'Submit',
    onClose: jest.fn(),
    onSubmit: jest.fn((e: React.FormEvent) => e.preventDefault()),
  }

  it('renders title, children and submit button when open', () => {
    render(
      <FormModal {...baseProps}>
        <input aria-label="field" />
      </FormModal>
    )
    expect(screen.getByRole('heading', { name: 'Create' })).toBeInTheDocument()
    expect(screen.getByLabelText('field')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Submit' })).toBeInTheDocument()
  })

  it('renders nothing when closed', () => {
    render(
      <FormModal {...baseProps} open={false}>
        <input aria-label="field" />
      </FormModal>
    )
    expect(screen.queryByRole('heading', { name: 'Create' })).not.toBeInTheDocument()
  })

  it('fires onSubmit when the form is submitted', async () => {
    const onSubmit = jest.fn((e: React.FormEvent) => e.preventDefault())
    render(
      <FormModal {...baseProps} onSubmit={onSubmit}>
        <input aria-label="field" />
      </FormModal>
    )
    await userEvent.click(screen.getByRole('button', { name: 'Submit' }))
    expect(onSubmit).toHaveBeenCalledTimes(1)
  })
})

