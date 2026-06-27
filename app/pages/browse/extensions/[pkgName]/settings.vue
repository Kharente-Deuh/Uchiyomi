<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<script setup lang="ts">
import type { PreferenceDto } from '~~/shared/dto/extensions'
import type { ExtensionSettingsDto } from '~~/shared/dto/extensions/extension-settings.dto'

const { pkgName } = useRoute().params
const { mobile } = useDisplay()

const { extension, settings, updateSettings } = useSingleExtension(pkgName)

const settingsCpy = ref<ExtensionSettingsDto>(settings.value ?? { common: [], sources: [] })
watch(settings, (value) => {
  if (value) {
    settingsCpy.value = value
  }
}, { once: true })

function setPreferenceValue(preference: PreferenceDto, value: unknown): void {
  switch (preference.type) {
    case 'switch':
    case 'checkbox':
      preference.booleanValue = value as boolean | undefined
      break
    case 'editText':
    case 'list':
      preference.textValue = value as string | undefined
      break
    case 'multiSelect':
      preference.multiValue = value as string[] | undefined
      break
  }
}

function updatePreference(key: string, sourceId: string | 'common', value: unknown): void {
  const preferences = sourceId === 'common'
    ? settingsCpy.value.common
    : settingsCpy.value.sources.find(s => s.id === sourceId)?.preferences

  const preference = preferences?.find(p => p.key === key)
  if (!preference) {
    return
  }

  setPreferenceValue(preference, value)
  updateSettings(settingsCpy.value)
}
</script>

<template>
  <OrganismPageLayout
    :title="extension?.name ?? ''"
    icon="fa6-solid:gear"
    :subtitle="$t('settings.title')"
    :show-back-route="true"
    :back-route="`/browse/extensions/${pkgName}`"
  >
    <div class="d-flex flex-column" :class="mobile ? 'ga-2' : 'ga-4'">
      <SettingsCard
        v-if="settingsCpy.common.length > 0"
        :title="$t('sources.settings.common.title')"
      >
        <SettingsCardItem
          v-for="(preference, i) in settingsCpy.common"
          :key="i"
          :title="preference.title ?? ''"
          :subtitle="preference.summary"
          show-subtitle
          :wrap="preference.type === 'multiSelect' || preference.type === 'list' || preference.type === 'editText'"
        >
          <ExtensionsSettingsItem
            :item="preference"
            @update:model-value="updatePreference(preference.key as string, 'common', $event)"
          />
        </SettingsCardItem>
      </SettingsCard>

      <SettingsCard
        v-for="(source, index) in settingsCpy.sources"
        :key="index"
        :title="source.name ?? ''"
      >
        <SettingsCardItem
          v-for="(preference, i) in source.preferences"
          :key="i"
          :title="preference.title ?? ''"
          :subtitle="preference.summary"
          show-subtitle
          :wrap="preference.type === 'multiSelect' || preference.type === 'list' || preference.type === 'editText'"
        >
          <ExtensionsSettingsItem
            :item="preference"
            @update:model-value="updatePreference(preference.key as string, source.id, $event)"
          />
        </SettingsCardItem>
      </SettingsCard>
    </div>
  </OrganismPageLayout>
</template>
