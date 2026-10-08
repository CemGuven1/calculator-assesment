import type { ErrorCode, Operation } from './types.ts'

const GENERIC = 'Something went wrong. Please try again.'

/** The user-facing message for each error code. */
const MESSAGES: Record<ErrorCode, string> = {
  DIVISION_BY_ZERO: 'You can’t divide by zero.',
  DOMAIN_ERROR: 'The result isn’t a real number.',
  OVERFLOW: 'The result is too large to calculate.',
  UNKNOWN_OPERATION: 'That operation isn’t supported.',
  INTERNAL: 'The calculator service ran into a problem. Please try again.',
  NETWORK_ERROR: 'Can’t reach the calculator service. Check your connection and try again.',
  TIMEOUT: 'The calculator service took too long to respond. Please try again.',
  BAD_RESPONSE: 'The calculator service is unavailable right now. Please try again.',
  // These mean the client sent a request the user can't fix: the inputs are
  // validated before sending, so INVALID_OPERANDS can only be a client bug.
  INVALID_OPERANDS: GENERIC,
  INVALID_REQUEST: GENERIC,
  PAYLOAD_TOO_LARGE: GENERIC,
  UNSUPPORTED_MEDIA_TYPE: GENERIC,
  METHOD_NOT_ALLOWED: GENERIC,
  NOT_FOUND: GENERIC,
}

/** Messages that read better when they mention the operation. */
const OPERATION_MESSAGES: Partial<Record<Operation, Partial<Record<ErrorCode, string>>>> = {
  power: {
    DIVISION_BY_ZERO: 'Zero can’t be raised to a negative power.',
    DOMAIN_ERROR: 'A negative number can’t be raised to a fractional power.',
  },
  sqrt: {
    DOMAIN_ERROR: 'You can’t take the square root of a negative number.',
  },
}

/**
 * Returns the message to show the user for an error code. Unrecognized codes,
 * for example from a newer backend, get a generic message.
 */
export function errorMessage(code: string, operation: Operation): string {
  if (!isErrorCode(code)) {
    return GENERIC
  }
  return OPERATION_MESSAGES[operation]?.[code] ?? MESSAGES[code]
}

function isErrorCode(code: string): code is ErrorCode {
  return Object.hasOwn(MESSAGES, code)
}
