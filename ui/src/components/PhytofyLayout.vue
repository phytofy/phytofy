<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-container class="pa-0 ma-0" fluid>
    <v-row>
      <v-col lg="4" md="8" sm="12" cols="12">
        <v-slider
          dense
          :label="`Elevation ${8 + 2 * $props.value.elevation}&quot;`"
          v-model="$props.value.elevation"
          min="0"
          max="7"
          step="1"
          color="#000000"
          track-color="grey--lighten"
        />
      </v-col>
      <v-col lg="1" md="2" sm="6" cols="6">
        <v-text-field
          dense
          label="Spacing"
          v-model="$props.value.spacing"
          type="number"
          :min="2"
          suffix='"'
        />
      </v-col>
      <v-col lg="1" md="2" sm="6" cols="6">
        <v-select
          dense
          label="Formation"
          v-model="$props.value.formation"
          :items="formationItems"
        />
      </v-col>
      <v-col lg="2" md="4" sm="4" cols="4">
        <v-select
          dense
          label="Channel"
          v-model="$props.value.channel"
          :items="channelItems"
        />
      </v-col>
      <v-col lg="2" md="4" sm="4" cols="4">
        <v-select
          dense
          label="Scale"
          v-model="$props.value.scale"
          :items="scaleItems"
        />
      </v-col>
      <v-col lg="2" md="4" sm="4" cols="4">
        <v-select
          dense
          label="Luminaire(s)"
          v-model="$props.value.luminaires"
          :items="luminairesItems"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<script lang="ts">
import Vue from "vue";
import { FORMATIONS, CHANNELS, SCALES, LUMINAIRES } from "../shared";

export default Vue.extend({
  name: "PhytofyLayout",

  props: {
    value: {
      type: Object,
      required: true,
    },
  },

  data: () => ({
    formationItems: FORMATIONS,
    channelItems: CHANNELS,
    scaleItems: SCALES,
    luminairesItems: LUMINAIRES,
  }),
});
</script>
