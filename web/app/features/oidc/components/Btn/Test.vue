<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { string } from 'yup'

const props = defineProps<{ issuerUrl: string }>()

const { testLoading, test } = useOidc()
const isValid = ref(false)
const showModal = ref(false)
const testResponse = ref<TestResponse>()

async function validate(): Promise<void> {
  try {
    await string().required().url().validate(props.issuerUrl)

    isValid.value = true
  } catch {
    isValid.value = false
  }
}

async function doTest(): Promise<void> {
  testResponse.value = await test(props.issuerUrl)
  if (testResponse.value) {
    showModal.value = true
  }
}

watch(showModal, (val: boolean) => {
  if (!val) {
    testResponse.value = undefined
  }
})

watch(() => props.issuerUrl, validate)
</script>

<template>
  <OidcModalTest
    v-model="showModal"
    :data="testResponse"
  />

  <VBtn
    variant="tonal"
    class="border-thin-secondary"
    color="secondary"
    icon="fa6-solid:vial"
    :disabled="!isValid"
    size="small"
    :loading="testLoading"
    @click="doTest"
  />
</template>
