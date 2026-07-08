import { act, renderHook } from '@testing-library/react'
import { useModal } from './useModal'

type Item = { id: number }

describe('useModal', () => {
  it('starts closed', () => {
    const { result } = renderHook(() => useModal<Item>())
    expect(result.current.modal).toBeNull()
  })

  it('openCreate defaults the item to null', () => {
    const { result } = renderHook(() => useModal<Item>())
    act(() => result.current.openCreate())
    expect(result.current.modal).toEqual({ mode: 'create', item: null })
  })

  it('openEdit / openDelete carry the item', () => {
    const { result } = renderHook(() => useModal<Item>())

    act(() => result.current.openEdit({ id: 1 }))
    expect(result.current.modal).toEqual({ mode: 'edit', item: { id: 1 } })

    act(() => result.current.openDelete({ id: 2 }))
    expect(result.current.modal).toEqual({ mode: 'delete', item: { id: 2 } })
  })

  it('close resets to null', () => {
    const { result } = renderHook(() => useModal<Item>())
    act(() => result.current.openCreate({ id: 9 }))
    act(() => result.current.close())
    expect(result.current.modal).toBeNull()
  })
})

