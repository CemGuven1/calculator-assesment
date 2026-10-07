// Types for the backend's REST API, as defined in docs/DESIGN.md.

/** The operations the backend supports. */
export type Operation = 'add' | 'subtract' | 'multiply' | 'divide' | 'power' | 'sqrt' | 'percentage'

/** Body of POST /api/v1/calculate. sqrt takes one operand; the others take two. */
export interface CalculateRequest {
  operation: Operation
  operands: number[]
}

/** Body of a successful (200) calculate response. */
export interface CalculateResponse {
  result: number
}

/** Body of every error response (4xx and 5xx). */
export interface ErrorResponse {
  error: {
    code: string
    message: string
  }
}

/** Codes the backend sends in ErrorResponse.error.code. */
export type ServerErrorCode =
  | 'INVALID_REQUEST'
  | 'UNKNOWN_OPERATION'
  | 'INVALID_OPERANDS'
  | 'DIVISION_BY_ZERO'
  | 'DOMAIN_ERROR'
  | 'OVERFLOW'
  | 'PAYLOAD_TOO_LARGE'
  | 'UNSUPPORTED_MEDIA_TYPE'
  | 'METHOD_NOT_ALLOWED'
  | 'NOT_FOUND'
  | 'INTERNAL'

/** Codes the client uses when there is no usable response from the backend. */
export type ClientErrorCode = 'NETWORK_ERROR' | 'BAD_RESPONSE'

export type ErrorCode = ServerErrorCode | ClientErrorCode

/** A failed calculation, ready to show to the user. */
export interface ApiError {
  /** An ErrorCode, or an unrecognized code from a newer backend. */
  code: string
  /** A user-friendly message, not the backend's technical one. */
  message: string
}

/** What calculate resolves to. It never rejects. */
export type CalculateOutcome = { ok: true; result: number } | { ok: false; error: ApiError }
