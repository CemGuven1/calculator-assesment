import type { Ref } from 'react'

interface NumberFieldProps {
  id: string
  label: string
  value: string
  error?: string
  onChange: (value: string) => void
  ref?: Ref<HTMLInputElement>
}

/**
 * A labelled text input for one operand. The validation message is linked
 * with aria-describedby, so screen readers read it along with the field.
 */
export function NumberField({ id, label, value, error, onChange, ref }: NumberFieldProps) {
  const errorId = `${id}-error`
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      {/* A text input rather than type="number", which reports partial input
          such as "1e" as empty, and rather than inputMode="decimal", which
          has no minus key on iOS. */}
      <input
        ref={ref}
        id={id}
        type="text"
        autoComplete="off"
        autoCapitalize="off"
        spellCheck={false}
        value={value}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : undefined}
        onChange={(event) => onChange(event.target.value)}
      />
      {error && (
        <p id={errorId} className="field-error">
          {error}
        </p>
      )}
    </div>
  )
}
