// SPDX-License-Identifier: AGPL-3.0-or-later
import type { ComputedRef, Ref } from 'vue'
import type { InferType, ObjectSchema } from 'yup'

// ---------------------------------------------------------------------------
// Path types — defined locally (no @tanstack/vue-form runtime dependency).
// Paths use dot-notation with numeric segments for array indices,
// matching the runtime getByPath/setByPath split on '.':
//   'address.city', 'items.0.name', 'items.0'
// ---------------------------------------------------------------------------

type Primitive = string | number | boolean | bigint | symbol | null | undefined

/** All valid dotted-path keys for T (intermediate + leaf nodes). */
type DeepKeys<T>
  = T extends Primitive | Date
    ? never
    : T extends Array<infer Item>
      ? `${number}` | (`${number}.${DeepKeys<Item> & string}`)
      : T extends object
        ? {
            [K in keyof T & string]:
              | K
              | (DeepKeys<T[K]> extends never ? never : `${K}.${DeepKeys<T[K]> & string}`)
          }[keyof T & string]
        : never

/** The type at dotted path P within T. */
type DeepValue<T, P extends string>
  = P extends `${infer Head}.${infer Tail}`
    ? Head extends keyof T
      ? DeepValue<T[Head], Tail>
      : T extends Array<infer Item>
        ? DeepValue<Item, Tail>
        : never
    : P extends keyof T
      ? T[P]
      : T extends Array<infer Item>
        ? Item
        : never

// ---------------------------------------------------------------------------

export type AnyObjectSchema = ObjectSchema<any>
type ValidateOn = 'blur' | 'change' | 'submit'

/** Props shape consumed directly by Vuetify input components (v-bind). */
export interface VuetifyFieldProps<T> {
  'modelValue': T
  'onUpdate:modelValue': (value: T) => void
  'onBlur': () => void
  'errorMessages': string[]
  'hideDetails': boolean | 'auto'
  'label'?: string
  'required'?: boolean
  'disabled'?: boolean
  'readonly'?: boolean
}

export interface FieldApi<T> {
  name: string
  value: Readonly<Ref<T>>
  errors: Readonly<Ref<string[]>>
  isTouched: Readonly<Ref<boolean>>
  isValidating: Readonly<Ref<boolean>>
  required: boolean
  label?: string
  handleChange: (value: T) => void
  handleBlur: () => void
  /** Vuetify-ready props — spread onto a v-text-field/v-select/etc. */
  props: ComputedRef<VuetifyFieldProps<T>>
}

export interface ArrayFieldApi<T> {
  fields: ComputedRef<Array<{ key: string, index: number }>>
  push: (value: T) => void
  remove: (index: number) => void
  move: (from: number, to: number) => void
}

export interface UseFormOptions<S extends AnyObjectSchema> {
  schema: S
  initialValues: NoInfer<InferType<S>>
  onSubmit?: (values: InferType<S>) => void | Promise<void>
  /**
   * Controls WHEN a field's errors become visible (touch timing).
   * Validation itself always runs eagerly on the full schema.
   *
   * - `'blur'` (default): field becomes touched on `handleBlur`.
   * - `'change'`: field becomes touched on first `handleChange`.
   * - `'submit'`: fields are NOT auto-touched on blur/change;
   *   only `handleSubmit` (touchAll) reveals errors.
   */
  validateOn?: ValidateOn
  asyncDebounceMs?: number
  scrollToError?: boolean
  ignoreFields?: Array<string>
  disabled?: Ref<boolean>
  readonly?: Ref<boolean>
}

export interface UseFormReturn<T> {
  field: <P extends DeepKeys<T>>(path: P) => FieldApi<DeepValue<T, P>>
  array: <P extends DeepKeys<T>>(path: P) => ArrayFieldApi<DeepValue<T, P> extends Array<infer I> ? I : never>
  values: ComputedRef<T>
  isValid: ComputedRef<boolean>
  isDirty: ComputedRef<boolean>
  handleSubmit: () => Promise<void>
  reset: (values?: T) => void
  /**
   * Replaces the entire server-error map (not a merge).
   * Callers should pass the full set of errors each time.
   * Each error also marks the corresponding field as touched so it becomes visible.
   */
  setServerErrors: (errors: Partial<Record<string, string | string[]>>) => void
}
