import type { Operation } from '../api/types.ts'
import { OPERATIONS } from '../calculator/operations.ts'

const OPERATION_ORDER = Object.keys(OPERATIONS) as Operation[]

interface OperationPickerProps {
  value: Operation
  onChange: (operation: Operation) => void
}

/**
 * Native radio buttons styled as a grid of buttons, so the browser provides
 * the keyboard behavior: Tab reaches the group, arrow keys change the choice.
 */
export function OperationPicker({ value, onChange }: OperationPickerProps) {
  return (
    <fieldset className="operations">
      <legend>Operation</legend>
      <div className="operation-grid">
        {OPERATION_ORDER.map((operation) => {
          const { label, symbol } = OPERATIONS[operation]
          return (
            <label key={operation} className="operation">
              <input
                type="radio"
                name="operation"
                value={operation}
                checked={operation === value}
                onChange={() => onChange(operation)}
              />
              <span className="operation-face">
                <span className="operation-symbol" aria-hidden="true">
                  {symbol}
                </span>
                <span className="operation-label">{label}</span>
              </span>
            </label>
          )
        })}
      </div>
    </fieldset>
  )
}
