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
