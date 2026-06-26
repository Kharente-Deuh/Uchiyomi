<script setup lang="ts">
const props = withDefaults(defineProps<{
  url: string
  size?: 'small' | 'default' | 'large'
}>(), {
  size: 'default',
})

const avatarSize = computed(() => {
  switch (props.size) {
    case 'small':
      return '48px'
    case 'large':
      return '80px'
    default:
      return '52px'
  }
})
</script>

<template>
  <div class="extension-avatar border-thin" :style="{ borderRadius: size === 'large' ? '16px' : '8px' }">
    <VImg
      :src="url"
      fill
      aspect-ratio="1/1"
      :width="avatarSize"
    />
  </div>
</template>

<style lang="scss" scoped>
.extension-avatar {
  overflow: hidden;

  // The source icons ship with a transparent margin; scale the image up
  // and clip the overflow so only the artwork shows.
  :deep(.v-img__img) {
    transform: scale(1.3);
  }
}
</style>
