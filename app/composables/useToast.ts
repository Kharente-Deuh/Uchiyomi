// SPDX-License-Identifier: AGPL-3.0-or-later
import type { Ref } from 'vue'
import { ref } from 'vue'

export interface ToastMessage {
  text: string
  color: string
}

export interface ToastComposable {
  messages: Ref<ToastMessage[]>
  success: (text: string) => void
  error: (text: string) => void
}

// Module-level queue shared across the app, bound to a single <VSnackbarQueue>
// in the default layout. Any component or composable can push a toast via
// success()/error(); VSnackbarQueue displays them one at a time and removes
// each from the queue as it is dismissed.
const messages = ref<ToastMessage[]>([])

export function useToast(): ToastComposable {
  function success(text: string): void {
    messages.value.push({ text, color: 'success' })
  }

  function error(text: string): void {
    messages.value.push({ text, color: 'error' })
  }

  return { messages, success, error }
}
