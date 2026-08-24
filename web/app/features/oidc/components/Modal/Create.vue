<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import { array, boolean, object, string } from 'yup'
import { slugFromDisplayName } from '~/features/oidc/utils/slug-from-display-name'
import { useForm } from '~/utils/forms/use-form'

const show = defineModel<boolean>({ required: true })

const { t } = useI18n()
const { loading, create } = useOidc()

async function onSubmit(values: CreateOidcProviderRequest): Promise<void> {
  await create(values)

  show.value = false
}

const slugDirty = ref(false)

const { field, handleSubmit, reset, isValid } = useForm({
  schema: object({
    displayName: string().required().label(t('settings.oidc.fields.displayName.label')),
    slug: string()
      .required()
      .max(64)
      .matches(/^[a-z0-9]+(?:-[a-z0-9]+)*$/)
      .label(t('settings.oidc.fields.slug.label')),
    issuerUrl: string().required().url().label(t('settings.oidc.fields.issuerUrl.label')),
    clientId: string().required().label(t('settings.oidc.fields.clienId.label')),
    clientSecret: string().required().label(t('settings.oidc.fields.clientSecret.label')),
    autoProvision: boolean().required().label(t('settings.oidc.fields.autoProvision.label')),
    scopes: array().of(string().required()).required().min(1).label(t('settings.oidc.fields.scopes.label')),
    usernameClaim: string().required().label(t('settings.oidc.fields.usernameClame.label')),
    roleClaim: string().defined().label(t('settings.oidc.fields.roleClaim.label')),
    adminValues: array().of(string().required()).defined().label(t('settings.oidc.fields.adminValues.label')),
    allowedValues: array().of(string().required()).defined().label(t('settings.oidc.fields.allowedValues.label')),
  }),
  initialValues: {
    displayName: '',
    slug: '',
    issuerUrl: '',
    clientId: '',
    clientSecret: '',
    autoProvision: false,
    scopes: [],
    usernameClaim: '',
    roleClaim: '',
    adminValues: [],
    allowedValues: [],
  },
  onSubmit,
})

const autoProvisionProps = computed(() => {
  const { 'onUpdate:modelValue': onUpdate, ...props } = field('autoProvision').props

  return { ...props, 'onUpdate:modelValue': (value: boolean | null) => onUpdate(value ?? false) }
})

watch(show, (val: boolean): void => {
  if (val) {
    slugDirty.value = false
    reset()
  }
})

watch(() => field('displayName').props.modelValue, (name: string) => {
  if (slugDirty.value) {
    return
  }
  field('slug').props['onUpdate:modelValue'](slugFromDisplayName(name))
})

watch(() => field('roleClaim').props.modelValue, (value: string) => {
  if (!value) {
    field('allowedValues').props['onUpdate:modelValue']([])
    field('adminValues').props['onUpdate:modelValue']([])
  }
})
</script>

<template>
  <OrganismModal
    v-model="show"
    :title="$t('settings.oidc.create.title')"
    prepend-icon="fa6-solid:plus"
    :is-form-complete="isValid"
    :loading
    :submit-text="$t('actions.create')"
    @submit="handleSubmit"
    @cancel="show = false"
  >
    <div class="d-flex ga-4 text-truncate w-100 align-center pb-2">
      <VIcon icon="fa6-regular:address-card" color="secondary" />

      <span class="text-title-large font-title text-truncate">{{ $t('settings.oidc.category.providerInfos.title') }}</span>
    </div>
    <VTextField v-bind="field('displayName').props" />
    <VTextField
      v-bind="field('slug').props"
      :messages="[$t('settings.oidc.fields.slug.hint')]"
      @update:model-value="slugDirty = true; field('slug').props['onUpdate:modelValue']($event ?? '')"
    />
    <VTextField v-bind="field('clientId').props" />
    <AtomInputPassword v-bind="field('clientSecret').props" />
    <VTextField v-bind="field('issuerUrl').props">
      <template #append>
        <OidcBtnTest :issuer-url="field('issuerUrl').props.modelValue" />
      </template>
    </VTextField>
    <VCheckbox
      density="compact"
      :messages="[$t('settings.oidc.fields.autoProvision.hint')]"
      v-bind="autoProvisionProps"
    />

    <div class="d-flex ga-4 text-truncate w-100 align-center py-2">
      <VIcon icon="fa6-solid:key" color="secondary" />

      <span class="text-title-large font-title text-truncate">{{ $t('settings.oidc.category.claims.title') }}</span>
    </div>

    <VCombobox v-bind="field('scopes').props" />
    <VTextField v-bind="field('usernameClaim').props" />
    <VTextField
      v-bind="field('roleClaim').props"
      clearable
      @update:model-value="field('roleClaim').props['onUpdate:modelValue']($event ?? '')"
    />
    <VCombobox
      v-bind="field('adminValues').props"
      :disabled="!field('roleClaim').props.modelValue"
    />
    <VCombobox
      v-bind="field('allowedValues').props"
      :disabled="!field('roleClaim').props.modelValue"
    />
  </OrganismModal>
</template>
