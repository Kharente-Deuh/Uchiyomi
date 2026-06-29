// SPDX-License-Identifier: AGPL-3.0-or-later

import type { Ref } from 'vue'
import type { InferType } from 'yup'
import type { AnyObjectSchema, ArrayFieldApi, FieldApi, UseFormOptions, UseFormReturn, VuetifyFieldProps } from './types'
import { computed, reactive, readonly, ref, toRaw, watch } from 'vue'
import { getFieldMeta, validateValues } from './schema'

function flatten(obj: Record<string, any>, prefix = ''): Record<string, true> {
  const out: Record<string, true> = {}
  for (const [key, value] of Object.entries(obj ?? {})) {
    const path = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      Object.assign(out, flatten(value, path))
    } else {
      out[path] = true
    }
  }

  return out
}

function getByPath(obj: Record<string, any>, path: string): any {
  let acc: any = obj
  for (const key of path.split('.')) {
    acc = acc?.[key]
  }

  return acc
}

/**
 * Build the Vuetify-ready props bag for a field (inlined adapter).
 *
 * Returns a `reactive` object — NOT a `ComputedRef` — so consumers can bind it
 * directly with `v-bind="form.field('x').props"` in a template. Vue only
 * auto-unwraps top-level refs from the setup state; a `ComputedRef` reached
 * through a method call (`form.field('x').props`) is NOT unwrapped, so its
 * props would silently never reach the component. A reactive object is a plain
 * (proxied) object and binds correctly regardless of nesting, while the
 * `computed()` getters inside keep every prop reactive.
 */
function toVuetifyProps<T>(
  field: Omit<FieldApi<T>, 'props'>,
  isFieldValid: () => boolean,
  formOptions?: { disabled?: Ref<boolean>, readonly?: Ref<boolean> },
): VuetifyFieldProps<T> {
  return reactive({
    'modelValue': computed(() => field.value.value),
    'onUpdate:modelValue': field.handleChange,
    'onBlur': field.handleBlur,
    'errorMessages': computed(() => field.errors.value),
    'hideDetails': computed(() => (field.errors.value.length > 0 ? false : 'auto')),
    'label': field.label,
    'required': field.required,
    'disabled': computed(() => formOptions?.disabled?.value),
    'readonly': computed(() => formOptions?.readonly?.value),
    // Required fields are marked; once a required field validates clean it also
    // gets `field-valid`. Optional fields carry neither marker.
    'class': computed(() => [
      field.required && 'field-required',
      field.required && isFieldValid() && 'field-valid',
    ].filter((c): c is string => typeof c === 'string')),
  }) as VuetifyFieldProps<T>
}

export function useForm<S extends AnyObjectSchema>(
  options: UseFormOptions<S>,
): UseFormReturn<InferType<S>> {
  const { schema, initialValues, validateOn = 'blur', asyncDebounceMs = 0 } = options

  // Native Vue reactive store — no form-library runtime dependency.
  const values = ref<Record<string, any>>(structuredClone(initialValues as Record<string, any>))
  const snapshot = ref<Record<string, any>>(structuredClone(initialValues as Record<string, any>))
  const errors = ref<Record<string, string[]>>({})
  const touched = ref<Record<string, boolean>>({})
  const validating = ref(false)

  // Monotonic run token prevents stale async validation from clobbering newer results.
  let runToken = 0

  async function runValidation(): Promise<void> {
    const token = ++runToken
    validating.value = true
    const result = await validateValues(schema, values.value)
    // Only commit if this is still the latest run (drop stale results).
    if (token === runToken) {
      errors.value = result
      validating.value = false
    }
  }

  // Debounce helper for value-watch triggered validation.
  let debounceTimer: ReturnType<typeof setTimeout> | undefined

  function scheduleValidation(): void {
    if (asyncDebounceMs > 0) {
      clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => {
        runValidation()
      }, asyncDebounceMs)
    } else {
      runValidation()
    }
  }

  // Initial validation + on every value change (full schema → cross-field).
  watch(values, scheduleValidation, { deep: true, immediate: true })

  // isValid flows from the async validation path (errors ref populated by runValidation).
  // Never use validateSync: it breaks async yup validators.
  const isValid = computed(() => Object.keys(errors.value).length === 0)

  const isDirty = computed(() => {
    const ignore = new Set(options.ignoreFields)
    const current = flatten(values.value)
    const base = flatten(snapshot.value)
    const keys = new Set([...Object.keys(current), ...Object.keys(base)])
    for (const key of keys) {
      if (ignore.has(key)) {
        continue
      }

      if (getByPath(values.value, key) !== getByPath(snapshot.value, key)) {
        return true
      }
    }

    return false
  })

  const serverErrors = ref<Record<string, string[]>>({})

  function reset(next?: Record<string, any>): void {
    const target = toRaw(next ?? snapshot.value)
    values.value = structuredClone(target)
    snapshot.value = structuredClone(target)
    touched.value = {}
    serverErrors.value = {}
  }

  function setByPath(obj: Record<string, any>, path: string, value: any): void {
    const keys = path.split('.')
    const last = keys.pop() as string
    let parent: any = obj
    for (const key of keys) {
      parent[key] ??= {}
      parent = parent[key]
    }

    parent[last] = value
  }

  function fieldErrors(path: string): string[] {
    return serverErrors.value[path] ?? errors.value[path] ?? []
  }

  function touchAll(): void {
    for (const key of Object.keys(flatten(values.value))) {
      touched.value[key] = true
    }
  }

  async function handleSubmit(): Promise<void> {
    touchAll()
    await runValidation()
    if (Object.keys(errors.value).length > 0) {
      if (options.scrollToError !== false && typeof document !== 'undefined') {
        document.querySelector('.v-input--error')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }

      return
    }

    await options.onSubmit?.(values.value)
  }

  /**
   * Replaces the entire server-error map (not a merge).
   * Callers should pass the full set of errors each time.
   * Each error also marks the corresponding field as touched so it becomes visible.
   */
  function setServerErrors(errs: Partial<Record<string, string | string[]>>): void {
    const next: Record<string, string[]> = {}
    for (const [key, value] of Object.entries(errs)) {
      if (value == null) {
        continue
      }

      next[key] = Array.isArray(value) ? value : [value]
      touched.value[key] = true
    }

    serverErrors.value = next
  }

  const fieldCache = new Map<string, FieldApi<any>>()

  /**
   * Access a field by dotted path. Item fields in arrays are addressed by position:
   * `form.field('items.0.name')` reads whatever is currently at index 0, not by item identity.
   * This means the binding changes when items are reordered (use `array().fields` for keys).
   * The returned field is Vuetify-ready: spread `field.props` onto a v-input component.
   */
  function field(path: string): FieldApi<any> {
    const cached = fieldCache.get(path)
    if (cached) {
      return cached
    }

    const meta = getFieldMeta(schema, path)
    const base: Omit<FieldApi<any>, 'props'> = {
      name: path,
      value: readonly(computed(() => getByPath(values.value, path))) as FieldApi<any>['value'],
      errors: readonly(computed(() => (touched.value[path] ? fieldErrors(path) : []))) as FieldApi<any>['errors'],
      isTouched: readonly(computed(() => touched.value[path] ?? false)) as FieldApi<any>['isTouched'],
      isValidating: readonly(validating),
      required: meta.required,
      label: meta.label,
      handleChange(value: any) {
        setByPath(values.value, path, value)
        delete serverErrors.value[path]
        // 'change' mode: touch on first handleChange.
        if (validateOn === 'change') {
          touched.value[path] = true
        }
        // 'blur' and 'submit' modes: no touch on change.
      },
      handleBlur() {
        // 'submit' mode: handleBlur does NOT auto-touch; only handleSubmit reveals errors.
        if (validateOn !== 'submit') {
          touched.value[path] = true
        }
      },
    }

    const api: FieldApi<any> = Object.assign(base, {
      props: toVuetifyProps(
        base,
        () => fieldErrors(path).length === 0,
        { disabled: options.disabled, readonly: options.readonly },
      ),
    })

    fieldCache.set(path, api)

    return api
  }

  const arrayCache = new Map<string, ArrayFieldApi<any>>()

  /**
   * Access an array field by dotted path.
   * Item fields are addressed by position: `form.field('items.0.name')` reads whatever
   * is currently at index 0. Positions shift on remove/move, so never cache item-level
   * field refs across mutations.
   */
  function array(path: string): ArrayFieldApi<any> {
    const cached = arrayCache.get(path)
    if (cached) {
      return cached
    }

    const list = (): any[] => (getByPath(values.value, path) as any[]) ?? []

    const api: ArrayFieldApi<any> = {
      fields: computed(() => list().map((_, index) => ({ key: `${path}.${index}`, index }))),
      push(value: any) {
        setByPath(values.value, path, [...list(), value])
      },
      remove(index: number) {
        setByPath(values.value, path, list().filter((_, i) => i !== index))
      },
      move(from: number, to: number) {
        const next = [...list()]
        const [moved] = next.splice(from, 1)
        next.splice(to, 0, moved)
        setByPath(values.value, path, next)
      },
    }

    arrayCache.set(path, api)

    return api
  }

  return {
    field,
    array,
    values: computed(() => values.value),
    isValid,
    isDirty,
    handleSubmit,
    reset,
    setServerErrors,
  } as UseFormReturn<InferType<S>>
}
