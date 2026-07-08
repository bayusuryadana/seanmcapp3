import { render, screen } from '@testing-library/react'
import { AppAlert } from './AppAlert'

describe('AppAlert', () => {
  it('renders nothing when not visible', () => {
    const { container } = render(<AppAlert alert={{ visible: false, text: 'hidden' }} />)
    expect(container).toBeEmptyDOMElement()
  })

  it('renders the message when visible', () => {
    render(<AppAlert alert={{ visible: true, text: 'something broke' }} />)
    expect(screen.getByRole('alert')).toHaveTextContent('something broke')
  })
})

