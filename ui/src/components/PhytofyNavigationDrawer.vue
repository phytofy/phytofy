<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-navigation-drawer
    v-bind:value="value"
    v-on:input="setDrawer"
    app
    clipped
    :mini-variant="true"
  >
    <v-list dense>
      <v-tooltip right>
        <template v-slot:activator="{ on, attrs }">
          <v-list-item
            to="/schedules"
            active-class="primary--text"
            ripple
            v-bind="attrs"
            v-on="on"
          >
            <v-list-item-action>
              <v-icon>mdi-calendar</v-icon>
            </v-list-item-action>
            <v-list-item-title>Schedules</v-list-item-title>
          </v-list-item>
        </template>
        <span>Schedules</span>
      </v-tooltip>
      <v-tooltip right>
        <template v-slot:activator="{ on, attrs }">
          <v-list-item
            to="/simulator"
            active-class="primary--text"
            ripple
            v-bind="attrs"
            v-on="on"
          >
            <v-list-item-action>
              <v-icon>mdi-map-outline</v-icon>
            </v-list-item-action>
            <v-list-item-title>Simulator</v-list-item-title>
          </v-list-item>
        </template>
        <span>Simulator</span>
      </v-tooltip>
    </v-list>
    <template v-slot:append>
      <v-list>
        <v-tooltip right>
          <template v-slot:activator="{ on, attrs }">
            <v-list-item
              to="/information"
              active-class="primary--text"
              ripple
              v-bind="attrs"
              v-on="on"
            >
              <v-list-item-action>
                <v-icon>mdi-information-outline</v-icon>
              </v-list-item-action>
              <v-list-item-title>Information</v-list-item-title>
            </v-list-item>
          </template>
          <span>Information</span>
        </v-tooltip>
      </v-list>
    </template>
  </v-navigation-drawer>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "PhytofyNavigationDrawer",

  props: {
    value: {
      type: Boolean,
      required: true,
    },
  },

  methods: {
    setDrawer(value: boolean) {
      this.$emit("input", value);
    },
  },
});
</script>
