<template>
  <v-main>
    <v-card class="d-flex flex-column ma-0 pa-0 fill-height" flat>
      <v-card class="mx-4 mt-2 mb-2 px-4 py-2" align="center">
        <PhytofyLayout v-model="parameters.layout" />
      </v-card>
      <v-card class="mx-4 mt-2 mb-4 px-4 py-2 flex-grow-1 d-flex align-stretch">
        <PhytofySimulation class="flex-grow-1" v-model="parameters" />
      </v-card>
    </v-card>
    <v-dialog v-model="caution" persistent width="30%">
      <v-card>
        <v-card-title>CAUTION</v-card-title>
        <v-card-text>
          <p align="justify">
            This is only a simulation which assumes idealized parameters and
            operation (some of which are mentioned in the product specification)
            - as such it is not a replacement for measurements with a PAR sensor
            in a controlled target environment.
          </p>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="caution = false">Dismiss</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-main>
</template>

<script lang="ts">
import Vue from "vue";
import {
  ELEVATION_DEFAULT,
  SPACING_DEFAULT,
  FORMATION_DEFAULT,
  CHANNEL_DEFAULT,
  SCALE_DEFAULT,
  LUMINAIRES_DEFAULT,
} from "../shared";
import PhytofyLayout from "@/components/PhytofyLayout.vue";
import PhytofySimulation from "@/components/PhytofySimulation.vue";

export default Vue.extend({
  name: "PhytofySimulator",

  components: {
    PhytofyLayout,
    PhytofySimulation,
  },

  data: () => ({
    parameters: {
      layout: {
        elevation: ELEVATION_DEFAULT,
        spacing: SPACING_DEFAULT,
        formation: FORMATION_DEFAULT,
        channel: CHANNEL_DEFAULT,
        scale: SCALE_DEFAULT,
        luminaires: LUMINAIRES_DEFAULT,
      },
    },
    caution: true,
  }),

  computed: {
    cautionVisible() {
      return !this.$store.state.eulaVisible && this.caution;
    },
  },
});
</script>
