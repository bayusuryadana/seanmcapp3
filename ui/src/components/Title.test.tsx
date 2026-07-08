import { render, screen } from '@testing-library/react'
import { Title } from './Title'

describe('Title', () => {
  it('renders its children as an h2', () => {
    render(<Title>My Section</Title>)
    const heading = screen.getByRole('heading', { level: 2 })
    expect(heading).toHaveTextContent('My Section')
  })
})

