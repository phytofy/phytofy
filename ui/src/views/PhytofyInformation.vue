<!-- Copyright (c) 2020 OSRAM; Licensed under the MIT license. -->

<template>
  <v-main>
    <v-sheet class="pa-4" fill-height>
      <h2>Product Information</h2>
      <p>
        <span style="font-weight: bold;">PHYTOFY&reg; RL</span>
        <br />
        <span>Version {{ version }}</span>
        <br />
        <span>© 2020, OSRAM GmbH. All rights reserved.</span>
        <br />
        <a
          href="https://www.osram.com/phytofy"
          target="_blank"
          rel="noopener noreferrer"
        >https://www.osram.com/phytofy</a>
        <br />
        <span>&#112;&#104;&#121;&#116;&#111;&#102;&#121;&#115;&#117;&#112;&#112;&#111;&#114;&#116;&#64;&#111;&#115;&#114;&#97;&#109;&#46;&#99;&#111;&#109;</span>
        <br />
      </p>
      <h2>Terms of Service &amp; Credits</h2>
      <p>
        <v-btn @click="open">End User License Agreement</v-btn>
      </p>
      <p>
        <v-dialog v-model="dialog" persistent scrollable>
          <template v-slot:activator="{ on, attrs }">
            <v-btn v-bind="attrs" v-on="on">Licensing Information</v-btn>
          </template>
          <v-card>
            <v-card-title>Licensing Information</v-card-title>
            <v-divider></v-divider>
            <v-card-text>
THIRD_PARTY_LICENSES
            </v-card-text>
            <v-divider></v-divider>
            <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn text @click="dialog = false">Dismiss</v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </p>
    </v-sheet>
  </v-main>
</template>

<script lang="ts">
import Vue from "vue";

export default Vue.extend({
  name: "PhytofyInformation",

  computed: {
    version() {
      return process.env.VUE_APP_VERSION;
    },
  },

  data: () => ({
    dialog: false,
  }),

  methods: {
    open() {
      this.$store.commit("setEulaVisible", true);
    },
  },
});
</script>
