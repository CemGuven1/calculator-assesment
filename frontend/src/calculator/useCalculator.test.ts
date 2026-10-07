import { act, renderHook, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { jsonResponse } from '../test/http.ts'
import { useCalculator } from './useCalculator.ts'

const fetchMock = vi.fn<typeof fetch>()

beforeEach(() => {
  fetchMock.mockReset()
  vi.stubGlobal('fetch', fetchMock)
  vi.stubEnv('VITE_API_BASE_URL', '')
})

function renderCalculator(a: string, b: string) {
  const hook = renderHook(() => useCalculator())
  act(() => {
    hook.result.current.changeInput('a', a)
    hook.result.current.changeInput('b', b)
  })
  return hook
}

describe('useCalculator', () => {
  it('sends the request and stores the result', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { result: 5 }))
    const { result } = renderCalculator('2', '3')

    act(() => result.current.submit())

    expect(result.current.state.status.kind).toBe('loading')
    await waitFor(() =>
      expect(result.current.state.status).toEqual({
        kind: 'success',
        calculation: { operation: 'add', operands: [2, 3], result: 5 },
      }),
    )
    expect(fetchMock).toHaveBeenCalledOnce()
  })

  it('uses the selected operation', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { result: 4 }))
    const { result } = renderCalculator('16', '')

    act(() => result.current.selectOperation('sqrt'))
    act(() => result.current.submit())

    await waitFor(() => expect(result.current.state.status.kind).toBe('success'))
    expect(fetchMock.mock.calls[0][1]?.body).toBe('{"operation":"sqrt","operands":[16]}')
  })

  it('stores the user-friendly error message', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(422, { error: { code: 'DIVISION_BY_ZERO', message: 'division by zero' } }),
    )
    const { result } = renderCalculator('1', '0')

    act(() => result.current.selectOperation('divide'))
    act(() => result.current.submit())

    await waitFor(() =>
      expect(result.current.state.status).toEqual({ kind: 'error', message: 'You can’t divide by zero.' }),
    )
  })

  it('does not call the backend when an input is invalid', () => {
    const { result } = renderCalculator('abc', '')

    act(() => result.current.submit())

    expect(result.current.state.fieldErrors).toEqual({
      a: 'Enter a valid number, like -3.5 or 1e3.',
      b: 'Enter a number.',
    })
    expect(fetchMock).not.toHaveBeenCalled()
  })

  it('aborts the request on unmount', () => {
    let signal: AbortSignal | null | undefined
    fetchMock.mockImplementation((_, init) => {
      signal = init?.signal
      return new Promise(() => {}) // never settles
    })
    const { result, unmount } = renderCalculator('2', '3')
    act(() => result.current.submit())

    unmount()

    expect(signal?.aborted).toBe(true)
  })
})
