<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-app-bar app clipped-left dense class="ma-0 pa-0">
    <v-app-bar-nav-icon @click.stop="$emit('input', !value);"></v-app-bar-nav-icon>
    <v-img max-width="90" max-height="18" contain :src="require('@/assets/logo.svg')" />
    <v-toolbar-title>
      <b>PHYTOFY&reg; RL</b>
    </v-toolbar-title>
  </v-app-bar>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "PhytofyAppBar",

  props: {
    value: {
      type: Boolean,
      required: true,
    },
  },
});
</script>
