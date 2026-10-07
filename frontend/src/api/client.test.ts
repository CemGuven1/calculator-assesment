import { beforeEach, describe, expect, it, vi } from 'vitest'
import { jsonResponse } from '../test/http.ts'
import { calculate } from './client.ts'

const fetchMock = vi.fn<typeof fetch>()

beforeEach(() => {
  fetchMock.mockReset()
  vi.stubGlobal('fetch', fetchMock)
  vi.stubEnv('VITE_API_BASE_URL', '')
})

describe('calculate', () => {
  it('POSTs the operation and operands as JSON', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { result: 2.5 }))
    const signal = new AbortController().signal

    await calculate('divide', [10, 4], signal)

    expect(fetchMock).toHaveBeenCalledExactlyOnceWith('/api/v1/calculate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{"operation":"divide","operands":[10,4]}',
      signal,
    })
  })

  it('sends one operand for sqrt', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { result: 4 }))

    await calculate('sqrt', [16])

    expect(fetchMock.mock.calls[0][1]?.body).toBe('{"operation":"sqrt","operands":[16]}')
  })

  it('uses same-origin requests when VITE_API_BASE_URL is unset', async () => {
    vi.stubEnv('VITE_API_BASE_URL', undefined)
    fetchMock.mockResolvedValue(jsonResponse(200, { result: 3 }))

    await calculate('add', [1, 2])

    expect(fetchMock.mock.calls[0][0]).toBe('/api/v1/calculate')
  })

  it.each([
    ['http://localhost:8080', 'http://localhost:8080/api/v1/calculate'],
    ['http://localhost:8080/', 'http://localhost:8080/api/v1/calculate'],
    ['https://calc.example/base//', 'https://calc.example/base/api/v1/calculate'],
  ])('uses VITE_API_BASE_URL %s', async (baseUrl, wantUrl) => {
    vi.stubEnv('VITE_API_BASE_URL', baseUrl)
    fetchMock.mockResolvedValue(jsonResponse(200, { result: 3 }))

    await calculate('add', [1, 2])

    expect(fetchMock.mock.calls[0][0]).toBe(wantUrl)
  })

  it('resolves to the result', async () => {
    fetchMock.mockResolvedValue(jsonResponse(200, { result: 0.30000000000000004 }))

    await expect(calculate('add', [0.1, 0.2])).resolves.toEqual({ ok: true, result: 0.30000000000000004 })
  })

  it('turns a backend error into a user-friendly message', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(422, { error: { code: 'DIVISION_BY_ZERO', message: 'division by zero' } }),
    )

    await expect(calculate('divide', [1, 0])).resolves.toEqual({
      ok: false,
      error: { code: 'DIVISION_BY_ZERO', message: 'You can’t divide by zero.' },
    })
  })

  it('uses the operation to word the message', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(422, {
        error: { code: 'DOMAIN_ERROR', message: 'result is not a real number: square root of a negative number' },
      }),
    )

    await expect(calculate('sqrt', [-4])).resolves.toEqual({
      ok: false,
      error: { code: 'DOMAIN_ERROR', message: 'You can’t take the square root of a negative number.' },
    })
  })

  it('reports a network failure', async () => {
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    await expect(calculate('add', [1, 2])).resolves.toEqual({
      ok: false,
      error: {
        code: 'NETWORK_ERROR',
        message: 'Can’t reach the calculator service. Check your connection and try again.',
      },
    })
  })

  it.each([
    // What the Vite dev proxy sends when the backend isn't running.
    ['an empty 502 from the dev proxy', () => new Response('', { status: 502, headers: { 'Content-Type': 'text/plain' } })],
    ['an HTML error page', () => new Response('<h1>Service Unavailable</h1>', { status: 503 })],
    ['a 200 without a numeric result', () => jsonResponse(200, { result: '5' })],
    ['a 200 with an error body', () => jsonResponse(200, { error: { code: 'INTERNAL', message: 'x' } })],
    ['an error status without the error body', () => jsonResponse(400, { message: 'bad' })],
    ['an error body with a missing message', () => jsonResponse(400, { error: { code: 'INVALID_REQUEST' } })],
    ['JSON null', () => jsonResponse(200, null)],
  ])('reports %s as a bad response', async (_, response) => {
    fetchMock.mockResolvedValue(response())

    await expect(calculate('add', [1, 2])).resolves.toEqual({
      ok: false,
      error: {
        code: 'BAD_RESPONSE',
        message: 'The calculator service is unavailable right now. Please try again.',
      },
    })
  })
})
