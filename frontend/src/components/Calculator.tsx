import { useEffect, useRef } from 'react'
import { OPERATIONS } from '../calculator/operations.ts'
import { useCalculator } from '../calculator/useCalculator.ts'
import { NumberField } from './NumberField.tsx'
import { OperationPicker } from './OperationPicker.tsx'
import { ResultPanel } from './ResultPanel.tsx'

/**
 * The calculator form. It holds no logic of its own: state and validation live
 * in useCalculator, and every result comes from the backend.
 */
export function Calculator() {
  const { state, selectOperation, changeInput, submit } = useCalculator()
  const { operation, input, fieldErrors, status } = state
  const labels = OPERATIONS[operation].operandLabels
  const loading = status.kind === 'loading'

  const firstInput = useRef<HTMLInputElement>(null)
  const secondInput = useRef<HTMLInputElement>(null)
  const focusErrorAfterSubmit = useRef(false)

  // After a submit, move focus to the first invalid field, so keyboard and
  // screen reader users land on the problem. This runs after the render that
  // links the field to its message.
  useEffect(() => {
    if (!focusErrorAfterSubmit.current) {
      return
    }
    focusErrorAfterSubmit.current = false
    if (fieldErrors.a) {
      firstInput.current?.focus()
    } else if (fieldErrors.b) {
      secondInput.current?.focus()
    }
  }, [fieldErrors])

  return (
    <form
      className="calculator"
      aria-busy={loading}
      onSubmit={(event) => {
        event.preventDefault()
        focusErrorAfterSubmit.current = true
        submit()
      }}
    >
      <OperationPicker value={operation} onChange={selectOperation} />

      <div className={labels.length === 2 ? 'fields fields-pair' : 'fields'}>
        <NumberField
          ref={firstInput}
          id="operand-a"
          label={labels[0]}
          value={input.a}
          error={fieldErrors.a}
          onChange={(value) => changeInput('a', value)}
        />
        {labels.length === 2 && (
          <NumberField
            ref={secondInput}
            id="operand-b"
            label={labels[1]}
            value={input.b}
            error={fieldErrors.b}
            onChange={(value) => changeInput('b', value)}
          />
        )}
      </div>

      {/* aria-disabled instead of disabled: a disabled button would drop
          keyboard focus, and the reducer already ignores repeat submits. */}
      <button type="submit" className="calculate" aria-disabled={loading}>
        {loading ? 'Calculating…' : 'Calculate'}
      </button>

      <ResultPanel status={status} />
    </form>
  )
}
