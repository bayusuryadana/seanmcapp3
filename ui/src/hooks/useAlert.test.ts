import { act, renderHook } from '@testing-library/react'
import { useAlert } from './useAlert'

describe('useAlert', () => {
  it('starts hidden', () => {
    const { result } = renderHook(() => useAlert())
    expect(result.current.alert).toEqual({ visible: false, text: '' })
  })

  it('showError makes it visible with text, clearAlert resets', () => {
    const { result } = renderHook(() => useAlert())

    act(() => result.current.showError('boom'))
    expect(result.current.alert).toEqual({ visible: true, text: 'boom' })

    act(() => result.current.clearAlert())
    expect(result.current.alert).toEqual({ visible: false, text: '' })
  })
})

