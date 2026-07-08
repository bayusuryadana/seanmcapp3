import { render, screen } from '@testing-library/react'
import { CellTypography } from './CellTypography'

describe('CellTypography', () => {
  it('renders children when not done (primary text)', () => {
    render(<CellTypography done={false}>Coffee</CellTypography>)
    expect(screen.getByText('Coffee')).toBeInTheDocument()
  })

  it('renders children when done (muted text)', () => {
    render(<CellTypography done={true}>Rent</CellTypography>)
    expect(screen.getByText('Rent')).toBeInTheDocument()
  })
})

