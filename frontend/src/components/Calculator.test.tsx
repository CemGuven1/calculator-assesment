import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { jsonResponse } from '../test/http.ts'
import { Calculator } from './Calculator.tsx'

const fetchMock = vi.fn<typeof fetch>()

beforeEach(() => {
  fetchMock.mockReset()
  vi.stubGlobal('fetch', fetchMock)
  vi.stubEnv('VITE_API_BASE_URL', '')
})

function renderCalculator() {
  const user = userEvent.setup()
  render(<Calculator />)
  return user
}

const field = (name: string) => screen.getByRole('textbox', { name })
const calculateButton = () => screen.getByRole('button', { name: 'Calculate' })

/** The JSON body of the nth request sent to the backend. */
function requestBody(n = 0): unknown {
  return JSON.parse(String(fetchMock.mock.calls[n][1]?.body))
}

describe('Calculator', () => {
  describe('a successful calculation', () => {
    it('sends the numbers to the backend and shows its result', async () => {
      fetchMock.mockResolvedValue(jsonResponse(200, { result: 5 }))
      const user = renderCalculator()

      await user.type(field('First number'), '2')
      await user.type(field('Second number'), '3')
      await user.click(calculateButton())

      expect(await screen.findByText('5')).toBeInTheDocument()
      expect(screen.getByRole('status')).toHaveTextContent('2 + 3 = 5')
      expect(requestBody()).toEqual({ operation: 'add', operands: [2, 3] })
    })

    it('shows the result the backend returned, rounded only for display', async () => {
      fetchMock.mockResolvedValue(jsonResponse(200, { result: 0.30000000000000004 }))
      const user = renderCalculator()

      await user.type(field('First number'), '0.1')
      await user.type(field('Second number'), '0.2')
      await user.click(calculateButton())

      expect(await screen.findByText('0.3')).toBeInTheDocument()
    })

    it('uses the selected operation and its field labels', async () => {
      fetchMock.mockResolvedValue(jsonResponse(200, { result: 30 }))
      const user = renderCalculator()

      await user.click(screen.getByRole('radio', { name: 'Percent' }))
      await user.type(field('Percentage (%)'), '15')
      await user.type(field('Of number'), '200')
      await user.click(calculateButton())

      expect(await screen.findByText('15% of 200 =')).toBeInTheDocument()
      expect(requestBody()).toEqual({ operation: 'percentage', operands: [15, 200] })
    })
  })

  describe('validation errors', () => {
    it('flags empty fields and does not call the backend', async () => {
      const user = renderCalculator()

      await user.click(calculateButton())

      for (const name of ['First number', 'Second number']) {
        expect(field(name)).toHaveAttribute('aria-invalid', 'true')
        expect(field(name)).toHaveAccessibleDescription('Enter a number.')
      }
      expect(fetchMock).not.toHaveBeenCalled()
    })

    it.each([
      ['abc', 'Enter a valid number, like -3.5 or 1e3.'],
      ['1.2.3', 'A number can only have one decimal point.'],
      ['1,5', 'Use a dot for decimals, like 1.5.'],
    ])('rejects %j with a clear message', async (text, message) => {
      const user = renderCalculator()

      await user.type(field('First number'), '4')
      await user.type(field('Second number'), text)
      await user.click(calculateButton())

      expect(field('Second number')).toHaveAccessibleDescription(message)
      expect(field('First number')).not.toHaveAttribute('aria-invalid')
      expect(fetchMock).not.toHaveBeenCalled()
    })

    it('moves focus to the first invalid field', async () => {
      const user = renderCalculator()

      await user.type(field('First number'), '4')
      await user.type(field('Second number'), 'x')
      await user.click(calculateButton())

      expect(field('Second number')).toHaveFocus()
    })

    it('clears a field’s error as soon as it is edited', async () => {
      const user = renderCalculator()
      await user.click(calculateButton())

      await user.type(field('First number'), '7')

      expect(field('First number')).not.toHaveAttribute('aria-invalid')
      expect(field('Second number')).toHaveAttribute('aria-invalid', 'true')
    })
  })

  describe('API errors', () => {
    it('shows a backend error in words the user understands', async () => {
      fetchMock.mockResolvedValue(jsonResponse(422, { error: { code: 'DIVISION_BY_ZERO', message: 'division by zero' } }))
      const user = renderCalculator()

      await user.click(screen.getByRole('radio', { name: 'Divide' }))
      await user.type(field('First number'), '1')
      await user.type(field('Second number'), '0')
      await user.click(calculateButton())

      expect(await screen.findByRole('alert')).toHaveTextContent('You can’t divide by zero.')
    })

    it('shows a network failure', async () => {
      fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))
      const user = renderCalculator()

      await user.type(field('First number'), '1')
      await user.type(field('Second number'), '2')
      await user.click(calculateButton())

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'Can’t reach the calculator service. Check your connection and try again.',
      )
    })

    it('removes the error once the user edits an input', async () => {
      fetchMock.mockResolvedValue(jsonResponse(422, { error: { code: 'OVERFLOW', message: 'result is out of range' } }))
      const user = renderCalculator()
      await user.type(field('First number'), '1e308')
      await user.type(field('Second number'), '10')
      await user.click(calculateButton())
      await screen.findByRole('alert')

      await user.type(field('Second number'), '0')

      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    })
  })

  describe('unary operations', () => {
    it('hides the second field for square root and sends one operand', async () => {
      fetchMock.mockResolvedValue(jsonResponse(200, { result: 4 }))
      const user = renderCalculator()

      await user.click(screen.getByRole('radio', { name: 'Square root' }))

      expect(screen.getAllByRole('textbox')).toHaveLength(1)
      await user.type(field('Number'), '16')
      await user.click(calculateButton())

      expect(await screen.findByText('√16 =')).toBeInTheDocument()
      expect(screen.getByText('4')).toBeInTheDocument()
      expect(requestBody()).toEqual({ operation: 'sqrt', operands: [16] })
    })

    it('ignores whatever is left in the hidden second field', async () => {
      fetchMock.mockResolvedValue(jsonResponse(200, { result: 3 }))
      const user = renderCalculator()
      await user.type(field('Second number'), 'not a number')

      await user.click(screen.getByRole('radio', { name: 'Square root' }))
      await user.type(field('Number'), '9')
      await user.click(calculateButton())

      expect(await screen.findByText('√9 =')).toBeInTheDocument()
      expect(requestBody()).toEqual({ operation: 'sqrt', operands: [9] })
    })

    it('shows the backend’s error for the square root of a negative number', async () => {
      fetchMock.mockResolvedValue(
        jsonResponse(422, {
          error: { code: 'DOMAIN_ERROR', message: 'result is not a real number: square root of a negative number' },
        }),
      )
      const user = renderCalculator()

      await user.click(screen.getByRole('radio', { name: 'Square root' }))
      await user.type(field('Number'), '-4')
      await user.click(calculateButton())

      expect(await screen.findByRole('alert')).toHaveTextContent(
        'You can’t take the square root of a negative number.',
      )
    })

    it('restores the second field when switching back', async () => {
      const user = renderCalculator()
      await user.type(field('Second number'), '8')

      await user.click(screen.getByRole('radio', { name: 'Square root' }))
      await user.click(screen.getByRole('radio', { name: 'Add' }))

      expect(field('Second number')).toHaveValue('8')
    })
  })

  describe('loading state', () => {
    it('shows progress and ignores repeated submits', async () => {
      let respond: (response: Response) => void = () => {}
      fetchMock.mockReturnValue(new Promise((resolve) => (respond = resolve)))
      const user = renderCalculator()
      await user.type(field('First number'), '2')
      await user.type(field('Second number'), '3')

      await user.click(calculateButton())

      const button = screen.getByRole('button', { name: 'Calculating…' })
      expect(button).toHaveAttribute('aria-disabled', 'true')
      expect(screen.getByRole('status')).toHaveTextContent('Calculating…')
      await user.click(button)
      expect(fetchMock).toHaveBeenCalledOnce()

      respond(jsonResponse(200, { result: 5 }))
      expect(await screen.findByText('5')).toBeInTheDocument()
      expect(calculateButton()).toHaveAttribute('aria-disabled', 'false')
    })
  })

  describe('keyboard', () => {
    it('works without a mouse: arrows pick the operation, Enter calculates', async () => {
      fetchMock.mockResolvedValue(jsonResponse(200, { result: 3 }))
      const user = renderCalculator()

      await user.tab()
      expect(screen.getByRole('radio', { name: 'Add' })).toHaveFocus()
      await user.keyboard('{ArrowRight}{ArrowRight}{ArrowRight}')
      expect(screen.getByRole('radio', { name: 'Divide' })).toBeChecked()

      await user.tab()
      expect(field('First number')).toHaveFocus()
      await user.keyboard('9')
      await user.tab()
      await user.keyboard('3{Enter}')

      expect(await screen.findByText('9 ÷ 3 =')).toBeInTheDocument()
      expect(requestBody()).toEqual({ operation: 'divide', operands: [9, 3] })
    })
  })
})
