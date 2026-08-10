// SPDX-License-Identifier: AGPL-3.0-or-later

import type { Ref } from 'vue'
import type { InferType } from 'yup'
import type { AnyObjectSchema, ArrayFieldApi, FieldApi, UseFormOptions, UseFormReturn, VuetifyFieldProps as VuetifyFieldProperties } from './types'
import { computed, reactive, readonly, ref, toRaw, unref, watch } from 'vue'
import { getFieldMeta, validateValues } from './schema'

function flatten(object: Record<string, any>, prefix = ''): Record<string, true> {
  const out: Record<string, true> = {}
  if (!object) {
    return out
  }

  for (const [key, value] of Object.entries(object)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      Object.assign(out, flatten(value, path))
    } else {
      out[path] = true
    }
  }

  return out
}

function cloneRaw<T>(value: T): T {
  const raw = toRaw(unref(value))
  if (Array.isArray(raw)) {
    return raw.map(item => cloneRaw(item)) as T
  }

  if (raw instanceof Date) {
    return new Date(raw) as T
  }

  if (raw && typeof raw === 'object' && Object.getPrototypeOf(raw) === Object.prototype) {
    const out: Record<string, any> = {}
    for (const [key, item] of Object.entries(raw)) {
      out[key] = cloneRaw(item)
    }

    return out as T
  }

  return raw as T
}

function getByPath(object: Record<string, any>, path: string): any {
  let accumulator: any = object
  for (const key of path.split('.')) {
    accumulator = accumulator?.[key]
  }

  return accumulator
}

function toVuetifyProperties<T>(
  field: Omit<FieldApi<T>, 'props'>,
  isFieldValid: () => boolean,
  formOptions?: { disabled?: Ref<boolean>, readonly?: Ref<boolean> },
): VuetifyFieldProperties<T> {
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
    'class': computed(() => [
      field.required && 'field-required',
      field.required && isFieldValid() && 'field-valid',
    ].filter((c): c is string => typeof c === 'string')),
  }) as VuetifyFieldProperties<T>
}

export function useForm<S extends AnyObjectSchema>(
  options: UseFormOptions<S>,
): UseFormReturn<InferType<S>> {
  const { schema, initialValues, validateOn = 'blur', asyncDebounceMs = 0 } = options

  const values = ref<Record<string, any>>(cloneRaw(initialValues as Record<string, any>))
  const snapshot = ref<Record<string, any>>(cloneRaw(initialValues as Record<string, any>))
  const errors = ref<Record<string, string[]>>({})
  const touched = ref<Record<string, boolean>>({})
  const validating = ref(false)

  let runToken = 0

  async function runValidation(): Promise<void> {
    const token = ++runToken
    validating.value = true
    const result = await validateValues(schema, values.value)
    if (token === runToken) {
      errors.value = result
      validating.value = false
    }
  }

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

  watch(values, scheduleValidation, { deep: true, immediate: true })

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
    const target = cloneRaw(next ?? snapshot.value)
    values.value = target
    snapshot.value = cloneRaw(target)
    touched.value = {}
    serverErrors.value = {}
  }

  function setByPath(object: Record<string, any>, path: string, value: any): void {
    const keys = path.split('.')
    const last = keys.pop() as string
    let parent: any = object
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
    const flattened = flatten(values.value)
    for (const key of Object.keys(flattened)) {
      touched.value[key] = true
    }
  }

  async function handleSubmit(): Promise<void> {
    touchAll()
    await runValidation()
    if (Object.keys(errors.value).length > 0) {
      if (typeof document !== 'undefined' && options.scrollToError !== false) {
        document.querySelector('.v-input--error')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
      }

      return
    }

    await options.onSubmit?.(values.value)
  }

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

  function field(path: string): FieldApi<any> {
    const cached = fieldCache.get(path)
    if (cached) {
      return cached
    }

    const meta = getFieldMeta(schema, path)
    const base: Omit<FieldApi<any>, 'props'> = {
      name: path,
      value: readonly(computed(() => getByPath(values.value, path))) as FieldApi<any>['value'],
      errors: readonly(computed(() => (touched.value[path] ?? false ? fieldErrors(path) : []))) as FieldApi<any>['errors'],
      isTouched: readonly(computed(() => touched.value[path] ?? false)) as FieldApi<any>['isTouched'],
      isValidating: readonly(validating),
      required: meta.required,
      label: meta.label,
      handleChange(value: any) {
        setByPath(values.value, path, value)
        delete serverErrors.value[path]
        if (validateOn === 'change') {
          touched.value[path] = true
        }
      },
      handleBlur() {
        if (validateOn !== 'submit') {
          touched.value[path] = true
        }
      },
    }

    const api: FieldApi<any> = Object.assign(base, {
      props: toVuetifyProperties(
        base,
        () => fieldErrors(path).length === 0,
        { disabled: options.disabled, readonly: options.readonly },
      ),
    })

    fieldCache.set(path, api)

    return api
  }

  const arrayCache = new Map<string, ArrayFieldApi<any>>()

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
        setByPath(values.value, path, list().filter((_, index_) => index_ !== index))
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
