<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { boolean, object, string } from 'yup'
import { useForm } from '~/utils/forms/use-form'

defineProps<{
  loading: boolean
}>()

const { t } = useI18n()
const { provider, update, updateLoading } = useOidcProvider()

type Form = Pick<UpdateOidcProviderRequest, 'displayName' | 'issuerUrl' | 'clientId' | 'autoProvision'>

async function onSubmit(values: Form): Promise<void> {
  if (!provider.value) {
    return
  }

  await update({ ...provider.value, ...values })
}

const { field, handleSubmit, reset, isValid } = useForm({
  schema: object({
    displayName: string().required().label(t('settings.oidc.fields.displayName.label')),
    issuerUrl: string().required().url().label(t('settings.oidc.fields.issuerUrl.label')),
    clientId: string().required().label(t('settings.oidc.fields.clienId.label')),
    autoProvision: boolean().required().label(t('settings.oidc.fields.autoProvision.label')),
  }),
  initialValues: {
    displayName: '',
    issuerUrl: '',
    clientId: '',
    autoProvision: false,
  },
  onSubmit,
})

watch(provider, (value?: OidcProvider) => {
  if (value) {
    reset({
      displayName: value.displayName,
      issuerUrl: value.issuerUrl,
      clientId: value.clientId,
      autoProvision: value.autoProvision,
    })
  }
}, { immediate: true })

const hasChanged = computed(() => {
  if (!provider.value) {
    return false
  }

  return provider.value.displayName !== field('displayName').props.modelValue
    || provider.value.issuerUrl !== field('issuerUrl').props.modelValue
    || provider.value.clientId !== field('clientId').props.modelValue
    || provider.value.autoProvision !== field('autoProvision').props.modelValue
})

const autoProvisionProps = computed(() => {
  const { 'onUpdate:modelValue': onUpdate, ...props } = field('autoProvision').props

  return { ...props, 'onUpdate:modelValue': (value: boolean | null) => onUpdate(value ?? false) }
})

const isFormValid = computed(() => isValid.value && hasChanged.value)
</script>

<template>
  <MoleculeCardComponent
    :disabled="loading || updateLoading"
    :loading
    :title="$t('settings.oidc.category.providerInfos.title')"
    icon="fa6-regular:address-card"
  >
    <div class="provider-informations-grid">
      <VTextField v-bind="field('displayName').props" />
      <VTextField v-bind="field('clientId').props" />
      <VTextField v-bind="field('issuerUrl').props" style="height: fit-content">
        <template #append>
          <OidcBtnTest :issuer-url="field('issuerUrl').props.modelValue" />
        </template>
      </VTextField>
      <VCheckbox
        density="compact"
        :messages="[$t('settings.oidc.fields.autoProvision.hint')]"
        v-bind="autoProvisionProps"
      />
    </div>

    <div class="w-100 mt-6 d-flex justify-end">
      <VBtn
        :text="$t('actions.save')"
        :loading="updateLoading"
        color="secondary"
        class="border-thin-secondary"
        :disabled="!loading && !isFormValid"
        @click="handleSubmit"
      />
    </div>
  </MoleculeCardComponent>
</template>

<style lang="scss">
.provider-informations-grid {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: repeat(2, 1fr);
}

@media (max-width: 800px) {
  .provider-informations-grid {
    grid-template-columns: 1fr;
  }
}
</style>
