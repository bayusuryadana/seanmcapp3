import { modalTitle } from './modal'

describe('modalTitle', () => {
  it('maps each mode to a label', () => {
    expect(modalTitle.create).toBe('Create')
    expect(modalTitle.edit).toBe('Edit')
    expect(modalTitle.delete).toBe('Delete')
  })
})

