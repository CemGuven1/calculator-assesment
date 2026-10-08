import { describe, expect, it } from 'vitest'
import { errorMessage } from './errorMessages.ts'
import type { Operation } from './types.ts'

const GENERIC = 'Something went wrong. Please try again.'

describe('errorMessage', () => {
  it.each<[code: string, operation: Operation, message: string]>([
    ['DIVISION_BY_ZERO', 'divide', 'You can’t divide by zero.'],
    ['DIVISION_BY_ZERO', 'power', 'Zero can’t be raised to a negative power.'],
    ['DOMAIN_ERROR', 'sqrt', 'You can’t take the square root of a negative number.'],
    ['DOMAIN_ERROR', 'power', 'A negative number can’t be raised to a fractional power.'],
    ['DOMAIN_ERROR', 'add', 'The result isn’t a real number.'],
    ['OVERFLOW', 'multiply', 'The result is too large to calculate.'],
    ['UNKNOWN_OPERATION', 'add', 'That operation isn’t supported.'],
    ['INTERNAL', 'add', 'The calculator service ran into a problem. Please try again.'],
    ['NETWORK_ERROR', 'add', 'Can’t reach the calculator service. Check your connection and try again.'],
    ['TIMEOUT', 'add', 'The calculator service took too long to respond. Please try again.'],
    ['BAD_RESPONSE', 'add', 'The calculator service is unavailable right now. Please try again.'],
  ])('%s for %s', (code, operation, message) => {
    expect(errorMessage(code, operation)).toBe(message)
  })

  // These mean the client sent a request the user can't fix. The inputs are
  // validated before sending, so INVALID_OPERANDS can only be a client bug.
  it.each([
    'INVALID_OPERANDS',
    'INVALID_REQUEST',
    'PAYLOAD_TOO_LARGE',
    'UNSUPPORTED_MEDIA_TYPE',
    'METHOD_NOT_ALLOWED',
    'NOT_FOUND',
  ])(
    '%s gets the generic message',
    (code) => {
      expect(errorMessage(code, 'add')).toBe(GENERIC)
    },
  )

  it.each(['SOMETHING_NEW', '', 'constructor', '__proto__'])('unrecognized code %j gets the generic message', (code) => {
    expect(errorMessage(code, 'add')).toBe(GENERIC)
  })
})
