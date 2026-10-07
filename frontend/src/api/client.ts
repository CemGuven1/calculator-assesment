import { errorMessage } from './errorMessages.ts'
import type {
  CalculateOutcome,
  CalculateRequest,
  CalculateResponse,
  ErrorResponse,
  Operation,
} from './types.ts'

const CALCULATE_PATH = '/api/v1/calculate'

/**
 * The backend's base URL, from VITE_API_BASE_URL. Empty (the default) means
 * same-origin requests, which go through the Vite proxy in development.
 */
function baseUrl(): string {
  return (import.meta.env.VITE_API_BASE_URL ?? '').replace(/\/+$/, '')
}

/**
 * Asks the backend to apply an operation to the operands. It never rejects:
 * every failure, including network errors, resolves to an outcome with a
 * user-friendly message. Pass a signal to cancel the request.
 */
export async function calculate(
  operation: Operation,
  operands: number[],
  signal?: AbortSignal,
): Promise<CalculateOutcome> {
  const request: CalculateRequest = { operation, operands }
  let response: Response
  try {
    response = await fetch(baseUrl() + CALCULATE_PATH, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(request),
      signal,
    })
  } catch {
    // fetch only rejects when no response arrived: the network failed or the
    // request was aborted.
    return failure('NETWORK_ERROR', operation)
  }

  const body = await readJson(response)
  if (response.ok && isCalculateResponse(body)) {
    return { ok: true, result: body.result }
  }
  if (!response.ok && isErrorResponse(body)) {
    return failure(body.error.code, operation)
  }
  // Not the contract's JSON, such as an error page from a proxy whose
  // backend is down.
  return failure('BAD_RESPONSE', operation)
}

function failure(code: string, operation: Operation): CalculateOutcome {
  return { ok: false, error: { code, message: errorMessage(code, operation) } }
}

async function readJson(response: Response): Promise<unknown> {
  try {
    return await response.json()
  } catch {
    return undefined
  }
}

function isCalculateResponse(body: unknown): body is CalculateResponse {
  return isObject(body) && typeof body.result === 'number'
}

function isErrorResponse(body: unknown): body is ErrorResponse {
  return (
    isObject(body) &&
    isObject(body.error) &&
    typeof body.error.code === 'string' &&
    typeof body.error.message === 'string'
  )
}

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
