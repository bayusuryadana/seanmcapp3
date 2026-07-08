import { render, screen } from '@testing-library/react'
import { Chart } from './Chart'

describe('Chart', () => {
  it('renders the Balance title with data', () => {
    render(<Chart data={[{ date: 202405, sum: 100 }, { date: 202406, sum: 250 }]} />)
    expect(screen.getByText('Balance')).toBeInTheDocument()
  })
})

